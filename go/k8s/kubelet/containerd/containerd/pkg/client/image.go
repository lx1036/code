package client

import (
	"context"
	"fmt"
	"github.com/containerd/containerd/diff"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/snapshots"
	"golang.org/x/sync/semaphore"
	"net/http"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/images"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/images/remotes/docker"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// UnpackOpt provides configuration for unpack
type UnpackOpt func(context.Context, *UnpackConfig) error

// RemoteContext is used to configure object resolutions and transfers with
// remote content stores and image providers.
type RemoteContext struct {
	// Resolver is used to resolve names to objects, fetchers, and pushers.
	// If no resolver is provided, defaults to Docker registry resolver.
	Resolver *docker.Resolver

	// Unpack is done after an image is pulled to extract into a snapshotter.
	// It is done simultaneously for schema 2 images when they are pulled.
	Unpack bool
	// UnpackOpts handles options to the unpack call.
	UnpackOpts []UnpackOpt

	// MaxConcurrentDownloads is the max concurrent content downloads for each pull.
	MaxConcurrentDownloads int

	// HandlerWrapper wraps the handler which gets sent to dispatch.
	// Unlike BaseHandlers, this can run before and after the built
	// in handlers, allowing operations to run on the descriptor
	// after it has completed transferring.
	HandlerWrapper func(images.Handler) images.Handler
}

func defaultRemoteContext() *RemoteContext {
	return &RemoteContext{
		Resolver: docker.NewResolver(docker.ResolverOptions{
			Client: http.DefaultClient,
		}),
	}
}

// RemoteOpt allows the caller to set distribution options for a remote
type RemoteOpt func(*Client, *RemoteContext) error

// Image describes an image used by containers
type Image struct {
	client *Client

	i        images.Image
	platform platforms.MatchComparer
}

// NewImageWithPlatform returns a client image object from the metadata image
func NewImageWithPlatform(client *Client, i images.Image, platform platforms.MatchComparer) *Image {
	return &Image{
		client:   client,
		i:        i,
		platform: platform,
	}
}

// Pull downloads the provided content into containerd's content store
// and returns a platform specific image object
func (c *Client) Pull(ctx context.Context, ref string, opts ...RemoteOpt) (*Image, error) {
	pullCtx := defaultRemoteContext()
	for _, o := range opts {
		if err := o(c, pullCtx); err != nil {
			return nil, err
		}
	}

	if pullCtx.Unpack {
		// unpacker only supports schema 2 image, for schema 1 this is noop.
		u, err := c.newUnpacker(ctx, pullCtx)
		if err != nil {
			return nil, fmt.Errorf("create unpacker: %w", err)
		}

		pullCtx.HandlerWrapper = func(f images.Handler) images.Handler {
			var (
				lock   sync.Mutex
				layers = map[digest.Digest][]ocispec.Descriptor{}
			)
			return images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
				children, err := f.Handle(ctx, desc)
				if err != nil {
					return children, err
				}

				switch desc.MediaType {
				case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
					var nonLayers []ocispec.Descriptor
					var manifestLayers []ocispec.Descriptor

					// Split layers from non-layers, layers will be handled after
					// the config
					for _, child := range children {
						if images.IsLayerType(child.MediaType) {
							manifestLayers = append(manifestLayers, child)
						} else {
							nonLayers = append(nonLayers, child)
						}
					}

					lock.Lock()
					for _, nl := range nonLayers {
						layers[nl.Digest] = manifestLayers
					}
					lock.Unlock()

					children = nonLayers
				case images.MediaTypeDockerSchema2Config, ocispec.MediaTypeImageConfig:
					lock.Lock()
					l := layers[desc.Digest]
					lock.Unlock()
					if len(l) > 0 {
						atomic.AddInt32(unpacks, 1)
						eg.Go(func() error {
							return u.unpack(uctx, rCtx, f, desc, l)
						})
					}
				}
				return children, nil
			})
		}
	}

	img, err := c.fetch(ctx, pullCtx, ref, 1)
	if err != nil {
		return nil, err
	}

	// NOTE(fuweid): unpacker defers blobs download. before create image
	// record in ImageService, should wait for unpacking(including blobs
	// download).
	if pullCtx.Unpack {
		if unpackEg != nil {
			if err := unpackEg.Wait(); err != nil {
				return nil, err
			}
		}
	}

	img, err = c.createNewImage(ctx, img)
	if err != nil {
		return nil, err
	}

	i := NewImageWithPlatform(c, img, pullCtx.PlatformMatcher)

	if pullCtx.Unpack {
		if unpacks == 0 {
			// Try to unpack is none is done previously.
			// This is at least required for schema 1 image.
			if err := i.Unpack(ctx, pullCtx.Snapshotter, pullCtx.UnpackOpts...); err != nil {
				return nil, errors.Wrapf(err, "failed to unpack image on snapshotter %s", pullCtx.Snapshotter)
			}
		}
	}

	return i, nil
}

func (c *Client) fetch(ctx context.Context, rCtx *RemoteContext, ref string, limit int) (images.Image, error) {

	name, desc, err := rCtx.Resolver.Resolve(ctx, ref)
	if err != nil {
		return images.Image{}, fmt.Errorf("failed to resolve reference %q: %w", ref, err)
	}

	fetcher, err := rCtx.Resolver.Fetcher(ctx, name)
	if err != nil {
		return images.Image{}, fmt.Errorf("failed to get fetcher for %q: %w", name, err)
	}

	// "application/vnd.docker.distribution.manifest.v2+json"

	// Get all the children for a descriptor
	childrenHandler := images.ChildrenHandler(store)
	// Set any children labels for that content
	childrenHandler = images.SetChildrenMappedLabels(store, childrenHandler, rCtx.ChildLabelMap)
	if rCtx.AllMetadata {
		// Filter manifests by platforms but allow to handle manifest
		// and configuration for not-target platforms
		childrenHandler = remotes.FilterManifestByPlatformHandler(childrenHandler, rCtx.PlatformMatcher)
	} else {
		// Filter children by platforms if specified.
		childrenHandler = images.FilterPlatforms(childrenHandler, rCtx.PlatformMatcher)
	}
	// Sort and limit manifests if a finite number is needed
	if limit > 0 {
		childrenHandler = images.LimitManifests(childrenHandler, rCtx.PlatformMatcher, limit)
	}

	appendDistSrcLabelHandler, err := docker.AppendDistributionSourceLabel(store, ref)
	if err != nil {
		return images.Image{}, err
	}

	handlers := append(rCtx.BaseHandlers,
		remotes.FetchHandler(store, fetcher),
		childrenHandler,
		appendDistSrcLabelHandler,
	)

	handler := images.Handlers(handlers...)

	if rCtx.HandlerWrapper != nil {
		handler = rCtx.HandlerWrapper(handler)
	}

	if rCtx.MaxConcurrentDownloads > 0 {
		limiter = semaphore.NewWeighted(int64(rCtx.MaxConcurrentDownloads))
	}

	if err := images.Dispatch(ctx, handler, limiter, desc); err != nil {
		return images.Image{}, err
	}

	return images.Image{
		Name:   name,
		Target: desc,
		Labels: rCtx.Labels,
	}, nil
}

const (
	// DefaultSnapshotter will set the default snapshotter for the platform.
	// This will be based on the client compilation target, so take that into
	// account when choosing this value.
	DefaultSnapshotter = "native" // darwin
	//DefaultSnapshotter = "overlayfs" // linux
)

// UnpackConfig provides configuration for the unpack of an image
type UnpackConfig struct {
	// ApplyOpts for applying a diff to a snapshotter
	ApplyOpts []diff.ApplyOpt
	// SnapshotOpts for configuring a snapshotter
	SnapshotOpts []snapshots.Opt
	// CheckPlatformSupported is whether to validate that a snapshotter
	// supports an image's platform before unpacking
	CheckPlatformSupported bool
}

type unpacker struct {
	updateCh    chan ocispec.Descriptor
	snapshotter string
	config      UnpackConfig
	c           *Client
	limiter     *semaphore.Weighted
}

func (c *Client) newUnpacker(ctx context.Context, rCtx *RemoteContext) (*unpacker, error) {
	var config UnpackConfig
	for _, o := range rCtx.UnpackOpts {
		if err := o(ctx, &config); err != nil {
			return nil, err
		}
	}
	var limiter *semaphore.Weighted
	if rCtx.MaxConcurrentDownloads > 0 {
		limiter = semaphore.NewWeighted(int64(rCtx.MaxConcurrentDownloads))
	}
	return &unpacker{
		updateCh:    make(chan ocispec.Descriptor, 128),
		snapshotter: DefaultSnapshotter,
		config:      config,
		c:           c,
		limiter:     limiter,
	}, nil
}
