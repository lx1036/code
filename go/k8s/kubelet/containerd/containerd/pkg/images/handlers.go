package images

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Handler handles image manifests
type Handler interface {
	Handle(ctx context.Context, desc ocispec.Descriptor) (subdescs []ocispec.Descriptor, err error)
}
