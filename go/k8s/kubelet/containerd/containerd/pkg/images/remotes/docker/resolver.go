package docker

import (
	"context"
	"errors"
	"fmt"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes/docker/schema1"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"strings"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/images/remotes/docker/reference"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// ErrInvalidAuthorization is used when credentials are passed to a server but
	// those credentials are rejected.
	ErrInvalidAuthorization = errors.New("authorization failed")

	// MaxManifestSize represents the largest size accepted from a registry
	// during resolution. Larger manifests may be accepted using a
	// resolution method other than the registry.
	//
	// NOTE: The max supported layers by some runtimes is 128 and individual
	// layers will not contribute more than 256 bytes, making a
	// reasonable limit for a large image manifests of 32K bytes.
	// 4M bytes represents a much larger upper bound for images which may
	// contain large annotations or be non-images. A proper manifest
	// design puts large metadata in subobjects, as is consistent the
	// intent of the manifest design.
	MaxManifestSize int64 = 4 * 1048 * 1048
)

// ResolverOptions are used to configured a new Docker register resolver
type ResolverOptions struct {
	// Hosts returns registry host configurations for a namespace.
	Hosts RegistryHosts

	// Client is the http client to used when making registry requests
	// Deprecated: use Hosts
	Client *http.Client
}

type Resolver struct {
	hosts         RegistryHosts
	header        http.Header
	resolveHeader http.Header
}

// NewResolver returns a new resolver to a Docker registry
func NewResolver(options ResolverOptions) *Resolver {

}

// Resolve ref="docker.io/library/nginx:1.17.8"
func (r *Resolver) Resolve(ctx context.Context, ref string) (string, ocispec.Descriptor, error) {
	base, err := r.resolveDockerBase(ref)
	if err != nil {
		return "", ocispec.Descriptor{}, err
	}
	spec := base.refspec
	if spec.Object == "" {
		return "", ocispec.Descriptor{}, reference.ErrObjectRequired
	}

	var paths [][]string
	caps := HostCapabilityPull
	digest := spec.Digest() // digest=""
	if digest != "" {
		if err := digest.Validate(); err != nil {
			// need to fail here, since we can't actually resolve the invalid
			// digest.
			return "", ocispec.Descriptor{}, err
		}

		// turns out, we have a valid digest, make a url.
		paths = append(paths, []string{"manifests", digest.String()})

		// fallback to blobs on not found.
		paths = append(paths, []string{"blobs", digest.String()})
	} else {
		// Add
		paths = append(paths, []string{"manifests", spec.Object})
		caps |= HostCapabilityResolve
	}

	hosts := base.filterHosts(caps)
	if len(hosts) == 0 {
		return "", ocispec.Descriptor{}, fmt.Errorf("no resolve hosts: not found")
	}
	var firstErr error
	for _, u := range paths {
		for _, host := range hosts {
			req := base.request(host, http.MethodHead, u...)
			if err := req.addNamespace(base.refspec.Hostname()); err != nil {
				return "", ocispec.Descriptor{}, err
			}
			for key, value := range r.resolveHeader {
				req.header[key] = append(req.header[key], value...)
			}

			resp, err := req.doWithRetries(ctx, nil)
			if err != nil {
				if errors.Is(err, ErrInvalidAuthorization) {
					err = fmt.Errorf("pull access denied, repository does not exist or may require authorization: %w", err)
				}
				// Store the error for referencing later
				if firstErr == nil {
					firstErr = err
				}

				klog.Errorf(fmt.Sprintf("trying next host: %+v", firstErr))
				continue // try another host
			}
			resp.Body.Close() // don't care about body contents.

			if resp.StatusCode > 299 {
				if resp.StatusCode == http.StatusNotFound {
					klog.Info("trying next host - response was http.StatusNotFound")
					continue
				}
				if resp.StatusCode > 399 {
					// Set firstErr when encountering the first non-404 status code.
					if firstErr == nil {
						firstErr = fmt.Errorf("pulling from host %s failed with status code %v: %v", host.Host, u, resp.Status)
					}
					continue // try another host
				}
				return "", ocispec.Descriptor{}, fmt.Errorf("pulling from host %s failed with unexpected status code %v: %v", host.Host, u, resp.Status)
			}
			size := resp.ContentLength
			contentType := getManifestMediaType(resp)

			// if no digest was provided, then only a resolve
			// trusted registry was contacted, in this case use
			// the digest header (or content from GET)
			if digest == "" {
				// this is the only point at which we trust the registry. we use the
				// content headers to assemble a descriptor for the name. when this becomes
				// more robust, we mostly get this information from a secure trust store.
				dgstHeader := digest.Digest(resp.Header.Get("Docker-Content-Digest"))

				if dgstHeader != "" && size != -1 {
					if err := dgstHeader.Validate(); err != nil {
						return "", ocispec.Descriptor{}, fmt.Errorf("%q in header not a valid digest: %w", dgstHeader, err)
					}
					digest = dgstHeader
				}
			}

			if digest == "" || size == -1 {
				klog.Info("no Docker-Content-Digest header, fetching manifest instead")

				req = base.request(host, http.MethodGet, u...)
				if err := req.addNamespace(base.refspec.Hostname()); err != nil {
					return "", ocispec.Descriptor{}, err
				}

				for key, value := range r.resolveHeader {
					req.header[key] = append(req.header[key], value...)
				}

				resp, err := req.doWithRetries(ctx, nil)
				if err != nil {
					return "", ocispec.Descriptor{}, err
				}
				defer resp.Body.Close()

				bodyReader := countingReader{reader: resp.Body}

				contentType = getManifestMediaType(resp)
				if digest == "" {
					if contentType == images.MediaTypeDockerSchema1Manifest {
						b, err := schema1.ReadStripSignature(&bodyReader)
						if err != nil {
							return "", ocispec.Descriptor{}, err
						}

						digest = digest.FromBytes(b)
					} else {
						digest, err = digest.FromReader(&bodyReader)
						if err != nil {
							return "", ocispec.Descriptor{}, err
						}
					}
				} else if _, err := io.Copy(io.Discard, &bodyReader); err != nil {
					return "", ocispec.Descriptor{}, err
				}
				size = bodyReader.bytesRead
			}
			// Prevent resolving to excessively large manifests
			if size > MaxManifestSize {
				if firstErr == nil {
					firstErr = fmt.Errorf("rejecting %d byte manifest for %s: not found", size, ref)
				}
				continue
			}

			desc := ocispec.Descriptor{
				Digest:    digest,
				MediaType: contentType,
				Size:      size,
			}

			klog.Infof(fmt.Sprintf("resolved desc.digest:%s", desc.Digest))
			return ref, desc, nil
		}
	}

	if firstErr == nil {
		firstErr = fmt.Errorf("%s: not found", ref)
	}

	return "", ocispec.Descriptor{}, firstErr
}

func (r *Resolver) Fetcher(ctx context.Context, ref string) (*Fetcher, error) {
	base, err := r.resolveDockerBase(ref)
	if err != nil {
		return nil, err
	}

	return &Fetcher{
		dockerBase: base,
	}, nil
}

func (r *Resolver) resolveDockerBase(ref string) (*dockerBase, error) {
	spec, err := reference.Parse(ref)
	if err != nil {
		return nil, err
	}

	host := spec.Hostname()
	hosts, err := r.hosts(host)
	if err != nil {
		return nil, err
	}

	return &dockerBase{
		refspec:    spec,
		repository: strings.TrimPrefix(spec.Locator, host+"/"),
		hosts:      hosts,
		header:     r.header,
	}, nil
}
