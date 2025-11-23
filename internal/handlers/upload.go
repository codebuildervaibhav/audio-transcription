package handlers

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/vaibh/audio-transcription/internal/queue"
	"github.com/vaibh/audio-transcription/internal/transcription"
	"github.com/vaibh/audio-transcription/internal/types"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	workerPool *queue.WorkerPool
	maxSizeMB  int
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(workerPool *queue.WorkerPool, maxSizeMB int) *UploadHandler {
	return &UploadHandler{
		workerPool: workerPool,
		maxSizeMB:  maxSizeMB,
	}
}

// Handle processes the upload request
func (h *UploadHandler) Handle(c *fiber.Ctx) error {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No file uploaded",
			"code":  "ERR_NO_FILE",
		})
	}

	// Get request name
	requestName := c.FormValue("name")
	if requestName == "" {
		requestName = "untitled"
	}

	// Validate file size
	maxSize := int64(h.maxSizeMB) * 1024 * 1024
	if file.Size > maxSize {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large (max %dMB)", h.maxSizeMB),
			"code":  "ERR_FILE_TOO_LARGE",
		})
	}

	// Validate file format
	if !transcription.ValidateAudioFormat(file.Filename) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported audio format",
			"code":  "ERR_INVALID_FORMAT",
		})
	}

	// Generate unique filename
	jobID := uuid.New().String()
	extension := filepath.Ext(file.Filename)
	tempPath := filepath.Join("temp", fmt.Sprintf("%s%s", jobID, extension))

	// Save file
	if err := c.SaveFile(file, tempPath); err != nil {
		log.Printf("Failed to save uploaded file: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save file",
			"code":  "ERR_SAVE_FAILED",
		})
	}

	// Create and enqueue job
	job := &queue.Job{
		ID:          jobID,
		RequestName: requestName,
		SourceType:  types.SourceUpload,
		FilePath:    tempPath,
	}

	h.workerPool.EnqueueJob(job)

	// Return job ID immediately
	return c.JSON(fiber.Map{
		"job_id":  jobID,
		"status":  "queued",
		"message": "File uploaded successfully, processing started",
	})
}
