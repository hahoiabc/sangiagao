package storage

import (
	"context"
	"io"
	"time"
)

// UploadResult holds the result of a file upload.
type UploadResult struct {
	Key string // object key, e.g. "listings/abc-def.jpg"
	URL string // public URL to access the file
}

// Client defines the interface for object storage operations.
type Client interface {
	Upload(ctx context.Context, folder, filename string, reader io.Reader, size int64, contentType string) (*UploadResult, error)
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
