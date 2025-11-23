package types

import "time"

// Job status constants
const (
	StatusQueued     = "QUEUED"
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

// Source type constants
const (
	SourceUpload  = "upload"
	SourceGDrive  = "gdrive"
	SourceYouTube = "youtube"
	SourceStream  = "stream"
)

// TranscriptionResult represents the output from Whisper
type TranscriptionResult struct {
	JobID       string
	Text        string
	Language    string
	Duration    float64
	Segments    []Segment
	WordCount   int
	ProcessedAt time.Time
	LocalPath   string
	GDriveURL   string
}

// Segment represents a timestamped segment of transcription
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
