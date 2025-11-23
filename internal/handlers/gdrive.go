package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/vaibh/audio-transcription/internal/queue"
	"github.com/vaibh/audio-transcription/internal/types"
)

// GDriveHandler handles Google Drive link processing
type GDriveHandler struct {
	workerPool *queue.WorkerPool
}

// NewGDriveHandler creates a new Google Drive handler
func NewGDriveHandler(workerPool *queue.WorkerPool) *GDriveHandler {
	return &GDriveHandler{
		workerPool: workerPool,
	}
}

// GDriveRequest represents the request body
type GDriveRequest struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Handle processes Google Drive link requests
func (h *GDriveHandler) Handle(c *fiber.Ctx) error {
	var req GDriveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "ERR_INVALID_BODY",
		})
	}

	// Validate URL
	if req.URL == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "URL is required",
			"code":  "ERR_NO_URL",
		})
	}

	// Extract file ID from various Google Drive URL formats
	fileID := extractGDriveFileID(req.URL)
	if fileID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Google Drive URL",
			"code":  "ERR_INVALID_URL",
		})
	}

	// Default name if not provided
	if req.Name == "" {
		req.Name = "gdrive_file"
	}

	// Generate job ID
	jobID := uuid.New().String()
	tempPath := filepath.Join("temp", fmt.Sprintf("%s.mp3", jobID))

	// Download file from Google Drive
	downloadURL := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", fileID)
	
	log.Printf("Downloading from Google Drive: %s", fileID)
	
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("Failed to download from Google Drive: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to download file from Google Drive",
			"code":  "ERR_DOWNLOAD_FAILED",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.Status(400).JSON(fiber.Map{
			"error": "File not accessible (may be private or doesn't exist)",
			"code":  "ERR_FILE_NOT_ACCESSIBLE",
		})
	}

	// Save to temp file
	out, err := os.Create(tempPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save downloaded file",
			"code":  "ERR_SAVE_FAILED",
		})
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to write downloaded file",
			"code":  "ERR_WRITE_FAILED",
		})
	}

	// Create and enqueue job
	job := &queue.Job{
		ID:          jobID,
		RequestName: req.Name,
		SourceType:  types.SourceGDrive,
		FilePath:    tempPath,
	}

	h.workerPool.EnqueueJob(job)

	return c.JSON(fiber.Map{
		"job_id":  jobID,
		"status":  "queued",
		"message": "Google Drive file downloaded, processing started",
	})
}

// extractGDriveFileID extracts the file ID from various Google Drive URL formats
func extractGDriveFileID(url string) string {
	// Pattern 1: https://drive.google.com/file/d/{ID}/view
	re1 := regexp.MustCompile(`/file/d/([a-zA-Z0-9_-]+)`)
	if matches := re1.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 2: https://drive.google.com/open?id={ID}
	re2 := regexp.MustCompile(`[?&]id=([a-zA-Z0-9_-]+)`)
	if matches := re2.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 3: Direct ID (25-40 characters)
	re3 := regexp.MustCompile(`^([a-zA-Z0-9_-]{25,40})$`)
	if matches := re3.FindStringSubmatch(url); len(matches) > 1 {
		return matches[1]
	}

	return ""
}
