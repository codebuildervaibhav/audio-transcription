# Audio-to-Text Backend - Project Context

## Project Overview

Building a Golang-based audio transcription backend that processes audio files, streams, Google Drive links, and YouTube videos (via headless browser). Transcripts are saved both locally and to Google Drive for analysis.

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | **Golang** |
| Web Framework | Fiber (fast HTTP framework) |
| AI Model | Whisper Small (via whisper.cpp Go bindings) |
| Audio Processing | FFmpeg |
| Database | SQLite (embedded, zero-config) |
| Queue | Go Channels (in-memory) |
| Cloud Storage | Google Drive API v3 |
| YouTube Capture | chromedp (headless Chrome) |
| WebSocket | Fiber WebSocket |

## System Flow & Architecture

### 1. Data Flow Pipeline
The application follows a linear pipeline architecture:

1.  **Input Layer**:
    *   **API**: Receives file uploads (`/upload`) or links (`/gdrive`, `/youtube`).
    *   **WebSocket**: Receives raw audio chunks (`/ws/stream`).
2.  **Handler Layer**:
    *   Validates input.
    *   Downloads content (if link) to `temp/`.
    *   Creates a `Job` object with a unique UUID.
    *   Enqueues the job into the `WorkerPool`.
3.  **Queue System**:
    *   Buffered Go Channels hold pending jobs.
    *   Decouples input (fast) from processing (slow).
4.  **Worker Layer**:
    *   4 concurrent workers pull jobs from the queue.
    *   **Normalization**: Converts any audio format to 16kHz WAV using `ffmpeg`.
    *   **Transcription**: Calls Python Whisper via subprocess.
5.  **Storage Layer**:
    *   **Local**: Saves `.txt` and `.json` metadata to `outputs/YYYY/MM/DD/`.
    *   **Cloud**: Uploads to Google Drive (if configured).
    *   **Database**: Records job metadata in SQLite.

### 2. Dependencies & "cu118" Explained

#### Core Dependencies
*   **Golang**: The orchestrator. Handles HTTP requests, concurrency, and file management.
    *   `gofiber/fiber`: High-performance web framework.
    *   `modernc.org/sqlite`: Pure Go SQLite driver (no CGO required).
*   **Python**: The AI engine.
    *   `openai-whisper`: The core transcription model.
    *   `torch` (PyTorch): The machine learning framework Whisper runs on.

#### What is "cu118"?
You will see references to `cu118` in the PyTorch installation (e.g., `torch --index-url .../cu118`).
*   **CUDA**: Compute Unified Device Architecture. It's NVIDIA's platform for parallel computing on GPUs.
*   **11.8**: The specific version of the CUDA Toolkit.
*   **Significance**: PyTorch must be built against the specific CUDA version installed on your system (or the one it bundles). Using the `cu118` version of PyTorch allows the application to offload the heavy matrix multiplications of the AI model to your **RTX 3060 GPU**, making transcription 10-50x faster than CPU.

#### External Tools
*   **FFmpeg**: The "Swiss Army Knife" of audio/video. Used to normalize inputs (e.g., convert a user's random MP3 or WebM file into the specific 16kHz WAV format Whisper requires).
*   **yt-dlp**: A command-line tool used to extract audio tracks from YouTube videos efficiently without downloading the video stream.

## Architecture Evolution Path

### NOW (Personal Use / MVP)
```
Go Channels Queue    →     (In-memory, 4 workers)
SQLite Database      →     (Single file, embedded)
Single Server        →     (All-in-one process)
```

### LATER (Multi-User / Production Scale)
```
Redis + Asynq Queue  →     (Distributed job queue)
PostgreSQL           →     (Scalable DB with better concurrency)
Multiple Workers     →     (Horizontal scaling across servers)
```

**Migration Path:** The architecture is designed to support this transition with minimal code changes.

## Input Methods Supported

1. **File Upload** (multipart/form-data)
2. **Google Drive Link** (public or authenticated)
3. **WebSocket Audio Stream** (live recording)
4. **YouTube Link** (headless browser capture - ToS compliant)

## Storage Strategy

### Dual Storage (Both Local + Google Drive)

Every successful transcription is saved to:

1. **Local Storage**: `/outputs/YYYY/MM/DD/{timestamp}_{requestName}.txt`
   - Fast access for analysis
   - Backup in case Drive upload fails
   - Includes metadata JSON file

2. **Google Drive**: `/Transcripts/YYYY/MM/DD/{timestamp}_{requestName}.txt`
   - Cloud backup
   - Accessible from any device
   - Shareable links
   - Includes metadata JSON file

**Metadata JSON Example:**
```json
{
  "job_id": "uuid-here",
  "request_name": "podcast_episode_42",
  "source_type": "youtube",
  "duration_seconds": 1847,
  "word_count": 3421,
  "model_used": "whisper-small",
  "language_detected": "en",
  "created_at": "2025-01-23T14:30:22Z",
  "gdrive_url": "https://drive.google.com/file/d/...",
  "segments": [
    {
      "start": 0.0,
      "end": 5.2,
      "text": "Welcome to the podcast..."
    }
  ]
}
```

## MVP Features (Phase 1)

- ✅ File upload transcription (MP3, WAV, M4A, etc.)
- ✅ Google Drive link processing
- ✅ WebSocket streaming
- ✅ YouTube headless capture
- ✅ Timestamps (word-level segments)
- ✅ Dual storage (local + Google Drive)
- ✅ Error handling & retry logic
- ✅ Temp file cleanup scheduler
- ✅ SQLite metadata tracking
- ❌ Speaker diarization (deferred to future phases)

## Configuration Overview

Key settings (stored in `config/config.yaml`):

```yaml
whisper:
  model: "whisper-small"       # Balance of speed/accuracy
  threads: 4                   # CPU cores for transcription

workers:
  count: 4                     # Concurrent job processors

storage:
  output_dir: "./outputs"      # Local transcript storage
  
google_drive:
  folder_name: "Transcripts"   # Drive folder structure

limits:
  max_file_size_mb: 500        # Prevent abuse
  max_duration_minutes: 120    # 2-hour limit
```

## Performance Expectations

- **Whisper Small Model**:
  - ~2GB RAM usage
  - ~20-30 seconds per minute of audio (on modern CPU)
  - Accuracy: ~95% for clear English audio

- **Concurrency**: 4 workers can process ~240 minutes/hour of audio

## Notes

- No speaker diarization in MVP (timestamps only)
- YouTube capture uses headless Chrome (ToS-compliant, no download)
- All transcripts saved to both local and Google Drive automatically
- SQLite handles up to 100K+ transcripts efficiently for personal use
