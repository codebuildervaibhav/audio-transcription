package queue

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/codebuildervaibhav/audio-transcription/internal/storage"
	"github.com/codebuildervaibhav/audio-transcription/internal/transcription"
	"github.com/codebuildervaibhav/audio-transcription/internal/types"
)

// WorkerPool manages a pool of workers processing transcription jobs
type WorkerPool struct {
	jobQueue     chan *Job
	workerCount  int
	transcriber  *transcription.WhisperTranscriber
	localStorage *storage.LocalStorage
	driveClient  *storage.DriveClient
	db           *storage.MetadataDB
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(
	workerCount int,
	transcriber *transcription.WhisperTranscriber,
	localStorage *storage.LocalStorage,
	driveClient *storage.DriveClient,
	db *storage.MetadataDB,
) *WorkerPool {
	return &WorkerPool{
		jobQueue:     make(chan *Job, 100), // Buffer of 100 jobs
		workerCount:  workerCount,
		transcriber:  transcriber,
		localStorage: localStorage,
		driveClient:  driveClient,
		db:           db,
	}
}

// Start initializes all workers
func (wp *WorkerPool) Start() {
	log.Printf("Starting worker pool with %d workers", wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(i)
	}
}

// EnqueueJob adds a job to the queue
func (wp *WorkerPool) EnqueueJob(job *Job) {
	job.Status = types.StatusQueued
	job.CreatedAt = time.Now()
	wp.jobQueue <- job
	log.Printf("Job %s enqueued (source: %s, name: %s)", job.ID, job.SourceType, job.RequestName)
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	log.Printf("Worker %d started", id)
	
	for job := range wp.jobQueue {
		// Panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Worker %d: PANIC processing job %s: %v\n%s", 
						id, job.ID, r, string(debug.Stack()))
					job.Status = types.StatusFailed
					job.Error = fmt.Errorf("Worker panic: %v", r)
					wp.cleanupTempFile(job.FilePath)
				}
			}()

			wp.processJob(id, job)
		}()
	}
}

// processJob handles the complete transcription pipeline
func (wp *WorkerPool) processJob(workerID int, job *Job) {
	log.Printf("Worker %d: Processing job %s", workerID, job.ID)
	job.Status = types.StatusProcessing

	// Step 1: Normalize audio
	normalizedPath, err := transcription.NormalizeAudio(job.FilePath)
	if err != nil {
		log.Printf("Worker %d: Audio normalization failed for job %s: %v", workerID, job.ID, err)
		job.Status = types.StatusFailed
		job.Error = fmt.Errorf("Audio normalization failed: %v", err)
		wp.cleanupTempFile(job.FilePath)
		return
	}
	defer wp.cleanupTempFile(normalizedPath)

	// Step 2: Transcribe with Whisper
	result, err := wp.transcriber.Transcribe(normalizedPath)
	if err != nil {
		log.Printf("Worker %d: Transcription failed for job %s: %v", workerID, job.ID, err)
		job.Status = types.StatusFailed
		job.Error = fmt.Errorf("Transcription failed: %v", err)
		wp.cleanupTempFile(job.FilePath)
		return
	}

	// Prepare result
	result.JobID = job.ID
	result.WordCount = len(strings.Fields(result.Text))
	result.ProcessedAt = time.Now()

	// Step 3: Save locally
	localPath, err := wp.localStorage.SaveTranscript(job.RequestName, result)
	if err != nil {
		log.Printf("Worker %d: Local save failed for job %s: %v", workerID, job.ID, err)
		job.Status = types.StatusFailed
		job.Error = fmt.Errorf("Local save failed: %v", err)
		wp.cleanupTempFile(job.FilePath)
		return
	}
	result.LocalPath = localPath

	// Step 4: Upload to Google Drive (with retry)
	var driveURL string
	if wp.driveClient != nil {
		for attempt := 1; attempt <= 3; attempt++ {
			driveURL, err = wp.driveClient.Upload(job.RequestName, result)
			if err == nil {
				result.GDriveURL = driveURL
				break
			}
			log.Printf("Worker %d: Google Drive upload attempt %d/3 failed: %v", workerID, attempt, err)
			if attempt < 3 {
				time.Sleep(time.Duration(attempt*attempt) * time.Second) // Exponential backoff
			}
		}
		if err != nil {
			log.Printf("Worker %d: WARNING - Google Drive upload failed after 3 attempts, continuing with local save only", workerID)
		}
	}

	// Step 5: Save metadata to database
	if wp.db != nil {
		err = wp.db.SaveTranscript(job.ID, job.RequestName, string(job.SourceType), 
			result.GDriveURL, localPath, result.Duration, result.WordCount)
		if err != nil {
			log.Printf("Worker %d: Database save failed: %v", workerID, err)
		}
	}

	// Step 6: Cleanup
	wp.cleanupTempFile(job.FilePath)

	job.Status = types.StatusCompleted
	log.Printf("Worker %d: Job %s completed successfully (local: %s, gdrive: %s)", 
		workerID, job.ID, localPath, driveURL)
}

// cleanupTempFile removes a temporary file
func (wp *WorkerPool) cleanupTempFile(filePath string) {
	if filePath == "" {
		return
	}
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to cleanup temp file %s: %v", filePath, err)
	}
}
