package handlers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/vaibh/audio-transcription/internal/queue"
	"github.com/vaibh/audio-transcription/internal/types"
)

// StreamHandler handles WebSocket audio streaming
type StreamHandler struct {
	workerPool *queue.WorkerPool
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(workerPool *queue.WorkerPool) *StreamHandler {
	return &StreamHandler{
		workerPool: workerPool,
	}
}

// Handle processes WebSocket connections
func (h *StreamHandler) Handle(c *websocket.Conn) {
	defer c.Close()

	var (
		buffer      bytes.Buffer
		requestName string
		jobID       = uuid.New().String()
	)

	log.Printf("WebSocket connection established: %s", jobID)

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Handle text messages (control)
		if messageType == websocket.TextMessage {
			msgStr := string(message)
			
			// Check for control messages
			if msgStr == "END" {
				log.Printf("Received END signal, processing stream...")
				break
			}
			
			// Set request name
			if len(msgStr) > 0 && len(msgStr) < 200 {
				requestName = msgStr
				log.Printf("Stream name set to: %s", requestName)
			}
			continue
		}

		// Handle binary messages (audio data)
		if messageType == websocket.BinaryMessage {
			buffer.Write(message)
		}
	}

	// If no data received, return
	if buffer.Len() == 0 {
		log.Printf("No audio data received in stream %s", jobID)
		return
	}

	// Default name if not set
	if requestName == "" {
		requestName = "stream_recording"
	}

	// Save buffered audio to temp file
	tempPath := filepath.Join("temp", fmt.Sprintf("%s.webm", jobID))
	
	if err := os.WriteFile(tempPath, buffer.Bytes(), 0644); err != nil {
		log.Printf("Failed to save stream buffer: %v", err)
		return
	}

	log.Printf("Stream saved to %s (%d bytes)", tempPath, buffer.Len())

	// Create and enqueue job
	job := &queue.Job{
		ID:          jobID,
		RequestName: requestName,
		SourceType:  types.SourceStream,
		FilePath:    tempPath,
	}

	h.workerPool.EnqueueJob(job)

	// Send confirmation
	c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"job_id":"%s","status":"queued"}`, jobID)))
}
