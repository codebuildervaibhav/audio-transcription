package transcription

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vaibh/audio-transcription/internal/types"
)

// WhisperTranscriber wraps Python's OpenAI Whisper for transcription
type WhisperTranscriber struct {
	modelName  string
	whisperCmd string
	threads    int
	mu         sync.Mutex // Thread-safe transcription
}

// NewWhisperTranscriber creates a new transcriber using Python Whisper
func NewWhisperTranscriber(modelPath string, threads int) (*WhisperTranscriber, error) {
	// For Python Whisper, we use the model name instead of path
	// Extract model name from path (e.g., "ggml-small.bin" -> "small")
	modelName := "small" // Default to small
	
	if strings.Contains(modelPath, "tiny") {
		modelName = "tiny"
	} else if strings.Contains(modelPath, "base") {
		modelName = "base"
	} else if strings.Contains(modelPath, "small") {
		modelName = "small"
	} else if strings.Contains(modelPath, "medium") {
		modelName = "medium"
	} else if strings.Contains(modelPath, "large") {
		modelName = "large"
	}

	log.Printf("Initializing Python Whisper with model: %s", modelName)
	log.Printf("Whisper will be called via: python -m whisper")
	log.Printf("Note: Whisper availability will be verified on first transcription")

	return &WhisperTranscriber{
		modelName:  modelName,
		whisperCmd: "python",
		threads:    threads,
	}, nil
}

// Transcribe processes an audio file and returns the transcript
func (wt *WhisperTranscriber) Transcribe(audioPath string) (*types.TranscriptionResult, error) {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	log.Printf("Transcribing with Python Whisper: %s", audioPath)

	// Create temp directory for Whisper output
	tempDir := filepath.Join("temp", "whisper_output")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir) // Clean up after

	// Get absolute path for audio file
	absAudioPath, err := filepath.Abs(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Python Whisper command using python -m whisper
	// Output formats: txt, json, srt, vtt, tsv
	cmd := exec.Command("python", "-m", "whisper",
		absAudioPath,
		"--model", wt.modelName,
		"--output_dir", tempDir,
		"--output_format", "json", // Get JSON for segments
		"--language", "en",         // Auto-detect if not specified
		"--fp16", "False",          // Disable fp16 for CPU compatibility
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper transcription failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("Whisper output: %s", string(output))

	// Read the JSON output file
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	jsonPath := filepath.Join(tempDir, baseName+".json")

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read whisper output: %v", err)
	}

	// Parse Whisper JSON output
	var whisperOutput WhisperOutput
	if err := json.Unmarshal(jsonData, &whisperOutput); err != nil {
		return nil, fmt.Errorf("failed to parse whisper JSON: %v", err)
	}

	// Convert to our format
	segments := make([]types.Segment, len(whisperOutput.Segments))
	for i, seg := range whisperOutput.Segments {
		segments[i] = types.Segment{
			Start: seg.Start,
			End:   seg.End,
			Text:  strings.TrimSpace(seg.Text),
		}
	}

	// Calculate duration (last segment end time)
	var duration float64
	if len(segments) > 0 {
		duration = segments[len(segments)-1].End
	}

	result := &types.TranscriptionResult{
		Text:     strings.TrimSpace(whisperOutput.Text),
		Language: whisperOutput.Language,
		Duration: duration,
		Segments: segments,
	}

	log.Printf("Transcription completed: %d segments, %.2fs duration", len(segments), duration)
	return result, nil
}

// WhisperOutput matches Python Whisper's JSON output format
type WhisperOutput struct {
	Text     string          `json:"text"`
	Language string          `json:"language"`
	Segments []WhisperSegment `json:"segments"`
}

// WhisperSegment represents a timestamped segment from Whisper
type WhisperSegment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
