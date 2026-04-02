package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig holds configuration for MinIO client.
type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
	PublicURL  string // base URL for public access, e.g. "http://localhost:9000"
}

// MinIOClient implements the Client interface using MinIO.
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

// NewMinIOClient creates a new MinIO storage client.
func NewMinIOClient(cfg MinIOConfig) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client init: %w", err)
	}

	return &MinIOClient{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}, nil
}

// EnsureBucket creates the bucket if it does not exist and sets public read policy.
func (m *MinIOClient) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	// Always ensure public read policy for images
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + m.bucketName + `/*"]}]}`
	if err := m.client.SetBucketPolicy(ctx, m.bucketName, policy); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}
	return nil
}

func (m *MinIOClient) Upload(ctx context.Context, folder, filename string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	key := folder + "/" + filename
	_, err := m.client.PutObject(ctx, m.bucketName, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload object: %w", err)
	}
	url := fmt.Sprintf("%s/%s", m.publicURL, key)
	return &UploadResult{Key: key, URL: url}, nil
}

func (m *MinIOClient) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}
	return url.String(), nil
}

func (m *MinIOClient) GetObject(ctx context.Context, key string) ([]byte, error) {
	obj, err := m.client.GetObject(ctx, m.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}
	return data, nil
}

func (m *MinIOClient) PresignedPutURL(ctx context.Context, folder, filename string, expiry time.Duration) (*PresignedPutResult, error) {
	key := folder + "/" + filename
	presignedURL, err := m.client.PresignedPutObject(ctx, m.bucketName, key, expiry)
	if err != nil {
		return nil, fmt.Errorf("presigned put url: %w", err)
	}
	// Rewrite internal MinIO URL to public nginx proxy path.
	// Internal: http://minio:9000/rice-images/listings/uuid.jpg?sig...
	// Public:   https://sangiagao.vn/storage/rice-images/listings/uuid.jpg?sig...
	// The nginx /storage/ location proxies PUT to minio:9000 with Host: minio:9000
	// so the presigned signature remains valid.
	internalBase := m.client.EndpointURL().String()
	// m.publicURL = "https://sangiagao.vn/images" → base = "https://sangiagao.vn"
	publicBase := strings.TrimSuffix(m.publicURL, "/images")
	publicBase = strings.TrimSuffix(publicBase, "/")
	uploadURL := strings.Replace(presignedURL.String(), strings.TrimSuffix(internalBase, "/"), publicBase+"/storage", 1)
	publicFileURL := fmt.Sprintf("%s/%s", m.publicURL, key)
	return &PresignedPutResult{
		UploadURL: uploadURL,
		PublicURL: publicFileURL,
		Key:       key,
	}, nil
}

func (m *MinIOClient) Delete(ctx context.Context, key string) error {
	return m.client.RemoveObject(ctx, m.bucketName, key, minio.RemoveObjectOptions{})
}
