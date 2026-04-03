package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sangiagao/rice-marketplace/pkg/storage"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

var (
	ErrInvalidFileType  = errors.New("chỉ chấp nhận file ảnh JPEG, PNG hoặc WebP")
	ErrInvalidAudioType = errors.New("chỉ chấp nhận file âm thanh AAC, MP4, OGG hoặc WAV")
	ErrFileTooLarge     = errors.New("file không được vượt quá 5MB")
)

const MaxImageSize = 5 * 1024 * 1024  // 5MB
const MaxAudioSize = 10 * 1024 * 1024 // 10MB

const thumbnailMaxDim = 600
const thumbnailJPEGQuality = 90

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var allowedAudioTypes = map[string]bool{
	"audio/aac":              true,
	"audio/mp4":              true,
	"audio/m4a":              true,
	"audio/x-m4a":            true,
	"audio/mpeg":             true,
	"audio/ogg":              true,
	"audio/wav":              true,
	"audio/x-wav":            true,
	"application/octet-stream": true,
}

// ImageUploadResult holds URLs for the original and thumbnail images.
type ImageUploadResult struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// PresignResult holds the presigned upload URL and the final public URL.
type PresignResult struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	Key       string `json:"key"`
}

// validMagicBytes checks the first bytes of file match JPEG, PNG, or WebP signature.
func validMagicBytes(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}
	// WebP: RIFF....WEBP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return true
	}
	return false
}

type UploadService struct {
	storage storage.Client
}

func NewUploadService(storageClient storage.Client) *UploadService {
	return &UploadService{storage: storageClient}
}

func (s *UploadService) UploadImage(ctx context.Context, folder string, file io.Reader, size int64, contentType, originalFilename string) (*ImageUploadResult, error) {
	if size > MaxImageSize {
		return nil, ErrFileTooLarge
	}

	if !allowedImageTypes[contentType] {
		return nil, ErrInvalidFileType
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}
	}

	// Read the entire file into memory so we can use it for both original upload and thumbnail generation.
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Validate file magic bytes match claimed content type
	if !validMagicBytes(fileBytes) {
		return nil, ErrInvalidFileType
	}

	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Upload original image.
	result, err := s.storage.Upload(ctx, folder, uniqueName, bytes.NewReader(fileBytes), int64(len(fileBytes)), contentType)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	uploadResult := &ImageUploadResult{
		URL: result.URL,
	}

	// Generate and upload thumbnail. If this fails, we still return the original URL.
	thumbURL, err := s.generateAndUploadThumbnail(ctx, folder, uniqueName, fileBytes)
	if err != nil {
		log.Printf("thumbnail generation failed for %s/%s: %v", folder, uniqueName, err)
	} else {
		uploadResult.ThumbnailURL = thumbURL
	}

	return uploadResult, nil
}

// generateAndUploadThumbnail decodes the image, resizes it to fit within 300x300
// preserving aspect ratio, encodes as JPEG, and uploads with a thumb_ prefix.
func (s *UploadService) generateAndUploadThumbnail(ctx context.Context, folder, originalName string, imgData []byte) (string, error) {
	src, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return "", fmt.Errorf("decoding image: %w", err)
	}

	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// Calculate new dimensions preserving aspect ratio.
	newW, newH := origW, origH
	if origW > thumbnailMaxDim || origH > thumbnailMaxDim {
		if origW >= origH {
			newW = thumbnailMaxDim
			newH = origH * thumbnailMaxDim / origW
		} else {
			newH = thumbnailMaxDim
			newW = origW * thumbnailMaxDim / origH
		}
	}

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: thumbnailJPEGQuality}); err != nil {
		return "", fmt.Errorf("encoding thumbnail: %w", err)
	}

	thumbName := "thumb_" + strings.TrimSuffix(originalName, filepath.Ext(originalName)) + ".jpg"

	thumbResult, err := s.storage.Upload(ctx, folder, thumbName, &buf, int64(buf.Len()), "image/jpeg")
	if err != nil {
		return "", fmt.Errorf("uploading thumbnail: %w", err)
	}

	return thumbResult.URL, nil
}

func (s *UploadService) GetPresignedPutURL(ctx context.Context, folder, contentType, ext string) (*PresignResult, error) {
	if !allowedImageTypes[contentType] {
		return nil, ErrInvalidFileType
	}
	if folder != "avatars" && folder != "listings" {
		folder = "images"
	}
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}
	}
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	result, err := s.storage.PresignedPutURL(ctx, folder, filename, 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("presign put url: %w", err)
	}
	return &PresignResult{
		UploadURL: result.UploadURL,
		PublicURL: result.PublicURL,
		Key:       result.Key,
	}, nil
}

// ConfirmPresignedUpload generates a thumbnail for an image uploaded via presigned URL.
// key is the object key, e.g. "listings/uuid.jpg"
func (s *UploadService) ConfirmPresignedUpload(ctx context.Context, key string) (string, error) {
	// Download original from MinIO
	imgData, err := s.storage.GetObject(ctx, key)
	if err != nil {
		return "", fmt.Errorf("download original: %w", err)
	}

	parts := strings.SplitN(key, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid key format: %s", key)
	}
	folder, filename := parts[0], parts[1]

	thumbURL, err := s.generateAndUploadThumbnail(ctx, folder, filename, imgData)
	if err != nil {
		return "", fmt.Errorf("generate thumbnail: %w", err)
	}
	return thumbURL, nil
}

func (s *UploadService) UploadAudio(ctx context.Context, file io.Reader, size int64, contentType, originalFilename string) (string, error) {
	if size > MaxAudioSize {
		return "", ErrFileTooLarge
	}

	if !allowedAudioTypes[contentType] {
		return "", ErrInvalidAudioType
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".m4a"
	}

	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	result, err := s.storage.Upload(ctx, "audio", uniqueName, file, size, contentType)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return result.URL, nil
}
