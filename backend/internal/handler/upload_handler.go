package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangiagao/rice-marketplace/internal/service"
)

type UploadHandler struct {
	uploadService UploadServiceInterface
}

func NewUploadHandler(uploadService UploadServiceInterface) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// UploadImage handles POST /upload/image (multipart form).
// Form fields: image (file, required), folder (string: "avatars" | "listings").
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	defer file.Close()

	folder := c.DefaultPostForm("folder", "images")
	if folder != "avatars" && folder != "listings" {
		folder = "images"
	}

	contentType := header.Header.Get("Content-Type")

	result, err := h.uploadService.UploadImage(
		c.Request.Context(),
		folder,
		file,
		header.Size,
		contentType,
		header.Filename,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFileType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrFileTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":           result.URL,
		"thumbnail_url": result.ThumbnailURL,
	})
}

// GetPresignedPutURL handles GET /upload/presign?folder=listings&content_type=image/jpeg&ext=.jpg
func (h *UploadHandler) GetPresignedPutURL(c *gin.Context) {
	folder := c.DefaultQuery("folder", "images")
	contentType := c.DefaultQuery("content_type", "image/jpeg")
	ext := c.DefaultQuery("ext", "")

	result, err := h.uploadService.GetPresignedPutURL(c.Request.Context(), folder, contentType, ext)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFileType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "presign failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": result.UploadURL,
		"public_url": result.PublicURL,
		"key":        result.Key,
	})
}

// ConfirmPresignedUpload handles POST /upload/confirm
// Generates thumbnail for an image uploaded via presigned URL.
func (h *UploadHandler) ConfirmPresignedUpload(c *gin.Context) {
	var req struct {
		Key string `json:"key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	thumbURL, err := h.uploadService.ConfirmPresignedUpload(c.Request.Context(), req.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "thumbnail generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"thumbnail_url": thumbURL})
}

func (h *UploadHandler) UploadAudio(c *gin.Context) {
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	url, err := h.uploadService.UploadAudio(
		c.Request.Context(),
		file,
		header.Size,
		contentType,
		header.Filename,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAudioType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrFileTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
