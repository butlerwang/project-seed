package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned when a key does not exist in storage.
var ErrNotFound = errors.New("storage: key not found")

// Storage is an S3-compatible object storage interface.
type Storage interface {
	Upload(ctx context.Context, key, contentType string, r io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}
