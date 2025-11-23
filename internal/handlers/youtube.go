package handlers

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/codebuildervaibhav/audio-transcription/internal/queue"
	"github.com/codebuildervaibhav/audio-transcription/internal/types"
)

// YouTubeHandler handles YouTube video audio capture
type YouTubeHandler struct {
	workerPool *queue.WorkerPool
}

// NewYouTubeHandler creates a new YouTube handler
func NewYouTubeHandler(workerPool *queue.WorkerPool) *YouTubeHandler {
	return &YouTubeHandler{
		workerPool: workerPool,
	}
}

// YouTubeRequest represents the request body
type YouTubeRequest struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Handle processes YouTube video requests
func (h *YouTubeHandler) Handle(c *fiber.Ctx) error {
	var req YouTubeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "ERR_INVALID_BODY",
		})
	}

	if req.URL == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "URL is required",
			"code":  "ERR_NO_URL",
		})
	}

	if req.Name == "" {
		req.Name = "youtube_video"
	}

	// Generate job ID
	jobID := uuid.New().String()
	tempPath := filepath.Join("temp", fmt.Sprintf("%s.opus", jobID))

	// Capture audio in background (this can take time for long videos)
	go func() {
		if err := h.captureYouTubeAudio(req.URL, tempPath); err != nil {
			log.Printf("Failed to capture YouTube audio: %v", err)
			return
		}

		// Create and enqueue job after capture completes
		job := &queue.Job{
			ID:          jobID,
			RequestName: req.Name,
			SourceType:  types.SourceYouTube,
			FilePath:    tempPath,
		}

		h.workerPool.EnqueueJob(job)
	}()

	return c.JSON(fiber.Map{
		"job_id":  jobID,
		"status":  "capturing",
		"message": "YouTube audio capture started (this may take a few minutes for long videos)",
	})
}

// captureYouTubeAudio uses headless Chrome to capture YouTube audio
func (h *YouTubeHandler) captureYouTubeAudio(url, outputPath string) error {
	// Create Chrome context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout for capture (max 30 minutes)
	ctx, cancel = context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	log.Printf("Starting YouTube capture: %s", url)

	// This is a simplified implementation
	// In production, you would use a more sophisticated approach:
	// 1. Extract audio URL from YouTube player
	// 2. Download audio stream directly
	// 3. Or use yt-dlp as a subprocess

	// For now, we'll use a workaround: extract the audio URL and download it
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second), // Wait for player to load
	)

	if err != nil {
		return fmt.Errorf("failed to navigate to YouTube: %v", err)
	}

	// Extract audio stream URL using JavaScript
	var audioURL string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			// This is a placeholder - actual implementation would need
			// to extract the audio stream URL from the YouTube player
			// For MVP, we recommend using yt-dlp as a subprocess instead
			"placeholder"
		`, &audioURL, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	)

	// NOTE: The headless Chrome approach for YouTube is complex
	// Recommended alternative: use yt-dlp subprocess
	// See captureWithYtDlp() below for a working implementation

	return h.captureWithYtDlp(url, outputPath)
}

// captureWithYtDlp uses yt-dlp to download YouTube audio (recommended)
func (h *YouTubeHandler) captureWithYtDlp(url, outputPath string) error {
	// Note: This requires yt-dlp to be installed
	// Install: pip install yt-dlp
	
	log.Printf("Using yt-dlp to download: %s", url)
	
	// Use yt-dlp to extract audio
	cmd := exec.Command("yt-dlp",
		"-x",                    // Extract audio
		"--audio-format", "opus", // Opus format
		"-o", outputPath,        // Output path
		url,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("YouTube audio downloaded successfully")
	return nil
}
