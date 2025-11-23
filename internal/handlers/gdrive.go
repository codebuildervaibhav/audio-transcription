package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	log.Printf("Downloading from Google Drive: %s", fileID)
	if err := downloadGDriveFile(fileID, tempPath); err != nil {
		log.Printf("Failed to download from Google Drive: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to download file: %v", err),
			"code":  "ERR_DOWNLOAD_FAILED",
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

// downloadGDriveFile handles the download logic including virus scan warnings
func downloadGDriveFile(fileID, destPath string) error {
	// 1. Try initial download with confirm=t (often works)
	url := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s&confirm=t", fileID)
	
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Check if we got an HTML page (likely a warning or login page) instead of the file
	contentType := resp.Header.Get("Content-Type")
	if len(contentType) >= 9 && contentType[:9] == "text/html" {
		// Read body to find confirmation token or error
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read warning page: %v", err)
		}
		bodyStr := string(body)

		// Check if it's a login page (file is private)
		if strings.Contains(bodyStr, "accounts.google.com") || strings.Contains(bodyStr, "signin") {
			return fmt.Errorf("file is private or not accessible (Google login required). Please make the file public ('Anyone with the link')")
		}
		
		// Look for confirm=XXXX pattern
		// Pattern: href="/uc?export=download&amp;id=...&amp;confirm=..."
		re := regexp.MustCompile(`confirm=([a-zA-Z0-9_-]+)`)
		matches := re.FindSubmatch(body)
		
		if len(matches) > 1 {
			token := string(matches[1])
			log.Printf("Found virus scan confirmation token: %s", token)
			
			// Retry with token
			url = fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s&confirm=%s", fileID, token)
			
			// Close previous response body before new request
			resp.Body.Close()
			
			resp, err = http.Get(url)
			if err != nil {
				return fmt.Errorf("failed to download with token: %v", err)
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != 200 {
				return fmt.Errorf("server returned status %d with token", resp.StatusCode)
			}
		} else {
			// Log a snippet of the body for debugging if token not found
			snippet := bodyStr
			if len(snippet) > 500 {
				snippet = snippet[:500]
			}
			log.Printf("HTML Response snippet: %s", snippet)
			return fmt.Errorf("received HTML response but could not find confirmation token (File might be private or format changed)")
		}
	}

	// Save to file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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
