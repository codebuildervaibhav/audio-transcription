# Audio-to-Text Backend Implementation Plan

## Overview

Building a production-ready audio transcription backend in **Golang** that processes audio files, streams, Google Drive links, and YouTube videos (via headless browser), with speaker diarization, automatic Google Drive backup, and scalable architecture.

## User Review Required

> [!IMPORTANT]
> **Google Drive Authentication**
> - Requires OAuth2 credentials (client ID + secret) from Google Cloud Console
> - Need to enable Google Drive API in your project
> - First run will require browser-based authentication flow
> - Credentials will be stored in `token.json` for future use

> [!WARNING]
> **YouTube Headless Approach**
> - Uses chromedp to launch headless Chrome and capture audio
> - Requires Chrome/Chromium installed on the server
> - Higher resource usage than simple download (but ToS-compliant)
> - May be slow for long videos (30+ min)

> [!NOTE]
> **Speaker Diarization - MVP Decision**
> - **Skipped in Phase 1**: MVP will include timestamps only, no speaker labels
> - Can be added in future phases via Python microservice (pyannote.audio) or cloud API
> - This simplifies the initial implementation while maintaining extensibility

---

## Proposed Changes

### Project Structure

```
audio-transcription-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point
├── internal/
│   ├── handlers/
│   │   ├── upload.go              # File upload handler
│   │   ├── gdrive.go              # Google Drive link handler
│   │   ├── stream.go              # WebSocket streaming
│   │   └── youtube.go             # YouTube headless capture
│   ├── transcription/
│   │   ├── whisper.go             # Whisper.cpp wrapper
│   │   ├── audio.go               # FFmpeg normalization
│   │   └── diarization.go         # Speaker diarization integration
│   ├── storage/
│   │   ├── gdrive_client.go       # Google Drive upload
│   │   ├── local.go               # Local file operations
│   │   └── metadata.go            # Database operations
│   ├── queue/
│   │   ├── worker.go              # Background job workers
│   │   └── jobs.go                # Job definitions
│   └── cleanup/
│       └── scheduler.go           # Temp file cleanup
├── outputs/                        # Local transcript storage
├── temp/                          # Temporary audio files
├── models/                        # Whisper models
├── config/
│   └── config.yaml                # Configuration
├── go.mod
└── README.md
```

---

### Core Components

#### 1. Web Framework & Server

**[NEW] [cmd/server/main.go](file:///cmd/server/main.go)**

- Use **Fiber** (fast HTTP framework for Go)
- Initialize routes: `/upload`, `/gdrive`, `/youtube`, `/ws/stream`
- Configure CORS for future frontend integration
- Startup cleanup of orphaned temp files
- Graceful shutdown handling

#### 2. Audio Processing

**[NEW] [internal/transcription/whisper.go](file:///internal/transcription/whisper.go)**

- Use `github.com/ggerganov/whisper.cpp/bindings/go` (official Go bindings)
- Load `whisper-small` model on initialization
- Thread-safe transcription with mutex locks
- Return structured output: text, segments (with timestamps), language detected

**[NEW] [internal/transcription/audio.go](file:///internal/transcription/audio.go)**

- FFmpeg wrapper to normalize all audio to 16kHz mono WAV
- Handles various input formats: MP3, M4A, WebM, OGG, etc.
- Returns path to normalized file
- Error handling for corrupted/unsupported files

**[NEW] [internal/transcription/diarization.go](file:///internal/transcription/diarization.go)**

- **MVP**: Placeholder returning empty speaker labels
- **Future**: HTTP client to call Python microservice running pyannote.audio
- Would accept WAV file, return JSON with speaker segments

---

#### 3. Input Handlers

**[NEW] [internal/handlers/upload.go](file:///internal/handlers/upload.go)**

- Accept multipart file upload
- Validate file type and size (max 500MB)
- Save to `/temp` with UUID filename
- Enqueue transcription job
- Return job ID immediately

**[NEW] [internal/handlers/gdrive.go](file:///internal/handlers/gdrive.go)**

- Parse Google Drive share URL (extract file ID)
- Download file stream directly to `/temp`
- Support both public and authenticated Drive files
- Enqueue transcription job

**[NEW] [internal/handlers/youtube.go](file:///internal/handlers/youtube.go)**

- Use `chromedp` to launch headless Chrome
- Navigate to YouTube URL
- Use `chrome.CaptureAudio()` to record audio stream
- Save captured audio to `/temp`
- Enqueue transcription job
- **Alternative**: Could use MediaRecorder API via CDP for better quality

**[NEW] [internal/handlers/stream.go](file:///internal/handlers/stream.go)**

- WebSocket endpoint accepting binary audio chunks
- Buffer incoming chunks into `bytes.Buffer`
- When client sends "END" message or disconnects:
  - Write buffer to temp WAV file
  - Enqueue transcription job
- Handle backpressure (if client sends data too fast)

---

#### 4. Job Queue (Scalability)

**[NEW] [internal/queue/worker.go](file:///internal/queue/worker.go)**

- Use **Go channels** for MVP (in-memory queue)
- Future: Redis-backed queue (Asynq library) for multi-server scaling
- Worker pool (configurable: default 4 workers)
- Each worker:
  1. Receives job from queue
  2. Normalizes audio
  3. Runs Whisper transcription
  4. Optionally runs diarization
  5. Saves to `/outputs` and Google Drive
  6. Updates database
  7. Cleans up temp file
- Panic recovery (worker restarts on crash)

**[NEW] [internal/queue/jobs.go](file:///internal/queue/jobs.go)**

- Job struct: `{ ID, RequestName, SourceType, FilePath, Status, CreatedAt }`
- Status states: QUEUED → PROCESSING → COMPLETED / FAILED

---

#### 5. Storage (Dual Save: Local + Google Drive)

> [!IMPORTANT]
> **Dual Storage for Analysis**
> - Every transcript is saved to BOTH local filesystem AND Google Drive
> - Local: Fast access, data analysis, backup if Drive fails
> - Google Drive: Cloud backup, accessible anywhere, shareable
> - Both locations include `.txt` transcript and `_meta.json` metadata

**[NEW] [internal/storage/gdrive_client.go](file:///internal/storage/gdrive_client.go)**

- OAuth2 authentication flow
- Upload transcript as `.txt` file
- Folder structure: `/Transcripts/YYYY/MM/DD/{timestamp}_{requestName}.txt`
- Also upload metadata JSON: `{timestamp}_{requestName}_meta.json`
- Return Google Drive file link

**[NEW] [internal/storage/local.go](file:///internal/storage/local.go)**

- Save transcript to `/outputs/{timestamp}_{requestName}.txt`
- Save metadata JSON: speaker segments, duration, model used, etc.
- Create dated subdirectories: `/outputs/2025/01/23/`

**[NEW] [internal/storage/metadata.go](file:///internal/storage/metadata.go)**

- SQLite database (simple, embedded, perfect for personal use + scalable to medium load)
- Table: `transcripts` (id, job_id, request_name, source_type, gdrive_url, local_path, created_at, duration, word_count)
- Insert on job completion
- Query endpoints for retrieving past transcripts

---

#### 6. Error Handling & Cleanup

**[NEW] [internal/cleanup/scheduler.go](file:///internal/cleanup/scheduler.go)**

- Runs on server startup: delete all files in `/temp` older than 24 hours
- Background ticker (every 1 hour): clean orphaned temp files
- Delete temp file immediately after successful transcription
- On error: mark temp file for retry (keep for 24h before deletion)

**Error Handling Strategy:**
- All handlers return structured JSON errors: `{ "error": "description", "code": "ERR_INVALID_FILE" }`
- Whisper crashes: Caught by worker panic recovery, job marked FAILED, temp file cleaned
- FFmpeg errors: Log error, mark job FAILED, notify via webhook (optional)
- Google Drive upload failure: Save locally, retry 3 times with exponential backoff

---

## Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Web Framework | [Fiber](https://gofiber.io/) | Fast, Express-like API, built on fasthttp |
| Whisper | [whisper.cpp Go bindings](https://github.com/ggerganov/whisper.cpp) | Native performance, no Python dependency |
| Audio Processing | FFmpeg (via exec) | Industry standard, handles all formats |
| Database | SQLite ([go-sqlite3](https://github.com/mattn/go-sqlite3)) | Embedded, zero-config, scales to 100K+ records |
| Queue (MVP) | Go Channels | Built-in, fast, good for single-server |
| Queue (Future) | [Asynq](https://github.com/hibiken/asynq) | Redis-backed, multi-server support |
| Google Drive | [Google Drive API v3](https://pkg.go.dev/google.golang.org/api/drive/v3) | Official SDK |
| YouTube Capture | [chromedp](https://github.com/chromedp/chromedp) | Headless Chrome automation |
| WebSocket | Fiber WebSocket | Built-in support |

---

## Verification Plan

### Automated Tests

1. **Unit Tests**
   - Audio normalization (various formats → WAV)
   - Whisper transcription mock
   - Google Drive upload mock

2. **Integration Tests**
   ```bash
   # File upload
   curl -F "file=@test.mp3" -F "name=TestAudio" http://localhost:3000/upload
   
   # Google Drive
   curl -X POST http://localhost:3000/gdrive \
     -H "Content-Type: application/json" \
     -d '{"url": "https://drive.google.com/file/d/...", "name": "DriveTest"}'
   
   # YouTube
   curl -X POST http://localhost:3000/youtube \
     -H "Content-Type: application/json" \
     -d '{"url": "https://www.youtube.com/watch?v=...", "name": "TED Talk"}'
   ```

3. **Load Testing**
   - Use `vegeta` to simulate 10 concurrent uploads
   - Verify worker pool handles queue correctly
   - Ensure no memory leaks with 100+ transcriptions

### Manual Verification

- [ ] Upload 5-minute podcast → Verify transcript accuracy
- [ ] Stream 30 seconds via WebSocket → Verify buffering works
- [ ] Test YouTube 10-minute video → Verify headless capture quality
- [ ] Check Google Drive folder structure created correctly
- [ ] Restart server mid-transcription → Verify cleanup works
- [ ] Delete `/temp` manually → Verify scheduler recreates directory

---

## Configuration File

**[NEW] [config/config.yaml](file:///config/config.yaml)**

```yaml
server:
  port: 3000
  host: "0.0.0.0"

whisper:
  model: "whisper-small"
  model_path: "./models/ggml-small.bin"
  threads: 4

workers:
  count: 4  # Number of concurrent transcription workers

storage:
  temp_dir: "./temp"
  output_dir: "./outputs"
  database: "./transcription.db"

cleanup:
  interval_minutes: 60
  max_age_hours: 24

google_drive:
  credentials_file: "./credentials.json"
  token_file: "./token.json"
  folder_name: "Transcripts"

limits:
  max_file_size_mb: 500
  max_duration_minutes: 120
```

---

## Next Steps

1. **Phase 1**: Core transcription (file upload + whisper + local storage)
2. **Phase 2**: Add WebSocket streaming + Google Drive integration
3. **Phase 3**: YouTube headless capture
4. **Phase 4**: Speaker diarization integration
5. **Phase 5**: Redis queue for horizontal scaling
