package client

import (
	"context"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/images"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// RemoteContext is used to configure object resolutions and transfers with
// remote content stores and image providers.
type RemoteContext struct {
}

// RemoteOpt allows the caller to set distribution options for a remote
type RemoteOpt func(*Client, *RemoteContext) error

// Image describes an image used by containers
type Image interface {
	// Name of the image
	Name() string
	// Target descriptor for the image content
	Target() ocispec.Descriptor
	// Labels of the image
	Labels() map[string]string
	// Unpack unpacks the image's content into a snapshot
	Unpack(context.Context, string, ...UnpackOpt) error
	// RootFS returns the unpacked diffids that make up images rootfs.
	RootFS(ctx context.Context) ([]digest.Digest, error)
	// Size returns the total size of the image's packed resources.
	Size(ctx context.Context) (int64, error)
	// Usage returns a usage calculation for the image.
	Usage(context.Context, ...UsageOpt) (int64, error)
	// Config descriptor for the image.
	Config(ctx context.Context) (ocispec.Descriptor, error)
	// IsUnpacked returns whether or not an image is unpacked.
	IsUnpacked(context.Context, string) (bool, error)
	// ContentStore provides a content store which contains image blob data
	//ContentStore() content.Store
	// Metadata returns the underlying image metadata
	Metadata() images.Image
}

// Pull downloads the provided content into containerd's content store
// and returns a platform specific image object
func (c *Client) Pull(ctx context.Context, ref string, opts ...RemoteOpt) (_ Image, retErr error) {
	pullCtx := defaultRemoteContext()
	for _, o := range opts {
		if err := o(c, pullCtx); err != nil {
			return nil, err
		}
	}

	if pullCtx.Unpack {

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

}
