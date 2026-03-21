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
