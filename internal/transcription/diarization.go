package transcription

// Speaker diarization module — identifies and segments speakers in audio.
// Planned integration with pyannote.audio or cloud APIs (AssemblyAI).

// DiarizationResult represents speaker segments (not implemented in MVP)
type DiarizationResult struct {
	Speakers []SpeakerSegment `json:"speakers"`
}

// SpeakerSegment represents when a speaker is talking
type SpeakerSegment struct {
	SpeakerID string  `json:"speaker_id"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
}

// PerformDiarization is a placeholder - returns empty result for MVP
func PerformDiarization(audioPath string) (*DiarizationResult, error) {
	// TODO: Implement speaker diarization in future phase
	// Options:
	// 1. Call Python microservice with pyannote.audio
	// 2. Use cloud API like AssemblyAI
	return &DiarizationResult{
		Speakers: []SpeakerSegment{},
	}, nil
}
