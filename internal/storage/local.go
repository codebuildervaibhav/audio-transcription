package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codebuildervaibhav/audio-transcription/internal/types"
)

// LocalStorage handles saving transcripts to the local filesystem
type LocalStorage struct {
	outputDir string
}

// NewLocalStorage creates a new local storage handler
func NewLocalStorage(outputDir string) *LocalStorage {
	return &LocalStorage{
		outputDir: outputDir,
	}
}

// SaveTranscript saves the transcript and metadata to local disk
func (ls *LocalStorage) SaveTranscript(requestName string, result *types.TranscriptionResult) (string, error) {
	// Create dated directory structure: outputs/2025/01/23/
	now := time.Now()
	dateDir := filepath.Join(ls.outputDir, 
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()))

	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create date directory: %v", err)
	}

	// Generate filename: 20250123_143022_podcast_episode.txt
	timestamp := now.Format("20060102_150405")
	baseFilename := fmt.Sprintf("%s_%s", timestamp, sanitizeFilename(requestName))
	
	txtPath := filepath.Join(dateDir, baseFilename+".txt")
	metaPath := filepath.Join(dateDir, baseFilename+"_meta.json")

	// Save transcript text
	if err := os.WriteFile(txtPath, []byte(result.Text), 0644); err != nil {
		return "", fmt.Errorf("failed to save transcript: %v", err)
	}

	// Save metadata JSON
	metadata := map[string]interface{}{
		"job_id":           result.JobID,
		"request_name":     requestName,
		"duration_seconds": result.Duration,
		"word_count":       result.WordCount,
		"model_used":       "whisper-small",
		"language":         result.Language,
		"created_at":       result.ProcessedAt,
		"segments":         result.Segments,
		"local_path":       txtPath,
		"gdrive_url":       result.GDriveURL,
	}

	metaJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(metaPath, metaJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to save metadata: %v", err)
	}

	return txtPath, nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscore
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for range invalid {
		result = filepath.Base(result) // Remove path separators
	}
	if len(result) > 100 {
		result = result[:100] // Limit length
	}
	return result
}
