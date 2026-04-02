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

// PresignedPutResult holds the presigned PUT URL and the final public URL.
type PresignedPutResult struct {
	UploadURL string // presigned PUT URL (client uploads here)
	PublicURL string // public URL after upload completes
	Key       string // object key
}

// Client defines the interface for object storage operations.
type Client interface {
	Upload(ctx context.Context, folder, filename string, reader io.Reader, size int64, contentType string) (*UploadResult, error)
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	PresignedPutURL(ctx context.Context, folder, filename string, expiry time.Duration) (*PresignedPutResult, error)
	GetObject(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
