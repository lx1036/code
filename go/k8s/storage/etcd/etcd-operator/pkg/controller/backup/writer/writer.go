package writer

import (
	"context"
	"io"
)

// Writer defines the required writer operations.
type Writer interface {
	// Write writes a backup file to the given path and returns size of written file.
	Write(ctx context.Context, path string, r io.Reader) (int64, error)

	// List backup files
	List(ctx context.Context, basePath string) ([]string, error)

	// Delete a backup file
	Delete(ctx context.Context, path string) error
}
