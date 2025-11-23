package transcription

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// NormalizeAudio converts any audio file to 16kHz mono WAV format
func NormalizeAudio(inputPath string) (string, error) {
	// Generate output path
	outputPath := filepath.Join("temp", fmt.Sprintf("normalized_%s.wav", uuid.New().String()))

	// FFmpeg command: convert to 16kHz mono WAV
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ar", "16000",      // 16kHz sample rate
		"-ac", "1",          // Mono
		"-c:a", "pcm_s16le", // 16-bit PCM
		"-y",                // Overwrite output
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\nOutput: %s", err, string(output))
	}

	return outputPath, nil
}

// ValidateAudioFormat checks if the file format is supported
func ValidateAudioFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := []string{".mp3", ".wav", ".m4a", ".ogg", ".flac", ".webm", ".aac", ".wma"}
	
	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}
