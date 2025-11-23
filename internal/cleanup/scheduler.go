package cleanup

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Scheduler handles cleanup of temporary files
type Scheduler struct {
	tempDir         string
	intervalMinutes int
	maxAgeHours     int
	stopChan        chan struct{}
}

// NewScheduler creates a new cleanup scheduler
func NewScheduler(tempDir string, intervalMinutes, maxAgeHours int) *Scheduler {
	return &Scheduler{
		tempDir:         tempDir,
		intervalMinutes: intervalMinutes,
		maxAgeHours:     maxAgeHours,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the cleanup scheduler
func (s *Scheduler) Start() {
	// Run initial cleanup on startup
	log.Println("Running initial temp file cleanup...")
	s.cleanOldFiles()

	// Start periodic cleanup
	ticker := time.NewTicker(time.Duration(s.intervalMinutes) * time.Minute)
	
	go func() {
		for {
			select {
			case <-ticker.C:
				s.cleanOldFiles()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	log.Printf("Cleanup scheduler started (interval: %dm, max age: %dh)", 
		s.intervalMinutes, s.maxAgeHours)
}

// Stop stops the cleanup scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
	log.Println("Cleanup scheduler stopped")
}

// cleanOldFiles removes files older than maxAgeHours from temp directory
func (s *Scheduler) cleanOldFiles() {
	now := time.Now()
	maxAge := time.Duration(s.maxAgeHours) * time.Hour
	
	var deletedCount int
	var deletedSize int64

	err := filepath.Walk(s.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file age
		age := now.Sub(info.ModTime())
		if age > maxAge {
			size := info.Size()
			if err := os.Remove(path); err != nil {
				log.Printf("Failed to delete old file %s: %v", path, err)
			} else {
				deletedCount++
				deletedSize += size
				log.Printf("Deleted old temp file: %s (age: %s, size: %dKB)", 
					filepath.Base(path), age.Round(time.Hour), size/1024)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	if deletedCount > 0 {
		log.Printf("Cleanup complete: %d files deleted, %.2fMB freed", 
			deletedCount, float64(deletedSize)/(1024*1024))
	}
}

// EnsureTempDirExists creates the temp directory if it doesn't exist
func EnsureTempDirExists(tempDir string) error {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}
	log.Printf("Temp directory ready: %s", tempDir)
	return nil
}
