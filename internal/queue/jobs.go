package queue

import (
	"time"

	"github.com/codebuildervaibhav/audio-transcription/internal/types"
)

// Job represents a transcription job
type Job struct {
	ID          string
	RequestName string
	SourceType  string
	FilePath    string
	Status      string
	Error       error
	Result      *types.TranscriptionResult
	CreatedAt   time.Time
}

// NewJob creates a new job with default values
func NewJob(id, requestName, sourceType, filePath string) *Job {
	return &Job{
		ID:          id,
		RequestName: requestName,
		SourceType:  sourceType,
		FilePath:    filePath,
		Status:      types.StatusQueued,
		CreatedAt:   time.Now(),
	}
}
