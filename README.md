# Audio Transcription Service

A high-performance, GPU-accelerated audio-to-text transcription service built with **Go** and **OpenAI Whisper**. Supports multiple input methods: file uploads, Google Drive links, YouTube videos, and real-time WebSocket audio streaming.

## Features

✅ **Multiple Input Methods**
- Direct file upload (MP3, WAV, M4A, OGG, FLAC, WebM, AAC, WMA)
- Google Drive links (public or authenticated)
- YouTube video audio extraction
- WebSocket real-time audio streaming

✅ **Dual Storage**
- Local filesystem (dated directory structure)
- Google Drive cloud backup (automatic folder hierarchy)

✅ **Production Ready**
- Concurrent worker pool for parallel processing
- SQLite metadata database
- Automatic temp file cleanup
- Panic recovery and error handling
- Graceful shutdown

✅ **Powered by OpenAI Whisper + CUDA GPU Acceleration**
- ~95% accuracy for clear English audio
- GPU-accelerated inference via PyTorch CUDA (cu118)
- Configurable model sizes (tiny, base, small, medium, large)
- Timestamps for each segment

---

## Prerequisites

### Required
1. **Go 1.21+** - [Download](https://golang.org/dl/)
2. **FFmpeg** - Audio format conversion
   ```bash
   # Windows (using Chocolatey)
   choco install ffmpeg
   
   # macOS
   brew install ffmpeg
   
   # Linux
   sudo apt install ffmpeg
   ```

3. **Python 3.9+** with **OpenAI Whisper** - AI transcription engine
   ```bash
   pip install openai-whisper
   
   # For GPU acceleration (recommended)
   pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118
   ```

### Optional
4. **yt-dlp** - For YouTube audio extraction and  Install a JavaScript runtime (required by yt-dlp for YouTube)
   ```bash
   pip install yt-dlp
   npm install -g esbuild  
   ```

5. **Google Drive API Credentials** - For cloud storage
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create new project or select existing
   - Enable **Google Drive API**
   - Create **OAuth 2.0 Client ID** (Desktop app)
   - Download `credentials.json` → place in project root

---

## Installation

### 1. Clone & Setup
```bash
git clone https://github.com/codebuildervaibhav/audio-transcription.git
cd audio-transcription

# Install Go dependencies
go mod download

# Verify FFmpeg is installed
ffmpeg -version

# Verify Whisper is installed
python -m whisper --help
```

### 2. Configure
Edit `config/config.yaml` if needed (defaults work for most cases):
```yaml
server:
  port: 3000
  host: "0.0.0.0"

whisper:
  model: "small"      # tiny, base, small, medium, large
  device: "cuda"       # cuda (GPU) or cpu

workers:
  count: 4             # Number of concurrent transcription workers
```

### 3. Run
```bash
# Development
go run ./cmd/server

# Production (build binary)
go build -o transcription-server ./cmd/server
./transcription-server
```

Server will start on `http://localhost:3000`

---

## API Usage

### 1. Upload Audio File
```bash
curl -F "file=@podcast.mp3" -F "name=MyPodcast" http://localhost:3000/upload
```

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "message": "File uploaded successfully, processing started"
}
```

### 2. Process Google Drive Link
```bash
curl -X POST http://localhost:3000/gdrive \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://drive.google.com/file/d/1AbC...XyZ/view",
    "name": "DriveRecording"
  }'
```

### 3. Extract YouTube Audio
```bash
curl -X POST http://localhost:3000/youtube \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "name": "TEDTalk"
  }'
```

### 4. WebSocket Streaming
```javascript
// Client-side JavaScript
const ws = new WebSocket('ws://localhost:3000/ws/stream');

// Send recording name
ws.send('LiveRecording');

// Send audio chunks (from MediaRecorder)
mediaRecorder.ondataavailable = (event) => {
  if (event.data.size > 0) {
    ws.send(event.data); // Send Blob as binary
  }
};

// Signal end of recording
ws.send('END');

// Receive job ID
ws.onmessage = (event) => {
  const response = JSON.parse(event.data);
  console.log('Job ID:', response.job_id);
};
```

### 5. List Transcripts
```bash
curl http://localhost:3000/transcripts
```

**Response:**
```json
[
  {
    "job_id": "...",
    "request_name": "MyPodcast",
    "source_type": "upload",
    "gdrive_url": "https://drive.google.com/file/d/.../view",
    "local_path": "./outputs/2025/01/23/20250123_143022_MyPodcast.txt",
    "created_at": "2025-01-23T14:30:22Z",
    "duration": 1847.5,
    "word_count": 3421
  }
]
```

---

## Output Structure

### Local Storage
```
outputs/
├── 2025/
│   └── 01/
│       └── 23/
│           ├── 20250123_143022_MyPodcast.txt       # Transcript text
│           └── 20250123_143022_MyPodcast_meta.json # Metadata
```

### Google Drive
```
Transcripts/
├── 2025/
│   └── 01/
│       └── 23/
│           ├── 20250123_143022_MyPodcast.txt
│           └── 20250123_143022_MyPodcast_meta.json
```

### Metadata JSON Example
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "request_name": "MyPodcast",
  "duration_seconds": 1847.5,
  "word_count": 3421,
  "model_used": "whisper-small",
  "language": "en",
  "created_at": "2025-01-23T14:30:22Z",
  "segments": [
    {
      "start": 0.0,
      "end": 5.2,
      "text": "Welcome to the podcast..."
    }
  ],
  "local_path": "./outputs/2025/01/23/20250123_143022_MyPodcast.txt",
  "gdrive_url": "https://drive.google.com/file/d/.../view"
}
```

---

## Troubleshooting

### Issue: "Whisper model not found"
**Solution:**
```bash
# Download whisper-small model
cd whisper.cpp
bash ./models/download-ggml-model.sh small
cp models/ggml-small.bin /path/to/listner/models/
```

### Issue: "ffmpeg: command not found"
**Solution:**
```bash
# Install FFmpeg
choco install ffmpeg  # Windows
brew install ffmpeg   # macOS
sudo apt install ffmpeg  # Linux
```

### Issue: "Google Drive upload failed"
**Solution:**
1. Check `credentials.json` exists in project root
2. Delete `token.json` and re-authenticate
3. Ensure Google Drive API is enabled in Cloud Console

### Issue: "YouTube capture failed"
**Solution:**
```bash
# Install yt-dlp
pip install yt-dlp

# Verify installation
yt-dlp --version
```

### Issue: "Database locked" error
**Solution:** SQLite doesn't handle concurrent writes well. Upgrade to PostgreSQL for production multi-server deployment (see [Scaling](#scaling)).

---

## Performance

### Whisper Small Model (CUDA GPU)
- **VRAM Usage:** ~2GB
- **Processing Speed:** ~5-10 seconds per minute of audio (NVIDIA GPU with CUDA)
- **Accuracy:** ~95% for clear English audio
- **Concurrency:** 4 workers = ~240 minutes/hour throughput

### Limits (Configurable in `config.yaml`)
- Max file size: 500MB
- Max duration: 120 minutes (2 hours)
- Worker pool: 4 concurrent jobs

---

## Scaling

### Current (Personal Use)
- Go Channels (in-memory queue)
- SQLite database
- Single server

### Future (Production)
To scale to multiple users/servers:

1. **Replace Go Channels with Redis + Asynq**
   ```go
   // Use github.com/hibiken/asynq
   // Allows multiple servers to share job queue
   ```

2. **Migrate SQLite to PostgreSQL**
   ```bash
   # Better concurrent write performance
   # Replace storage.NewMetadataDB with PostgreSQL driver
   ```

3. **Deploy Multiple Workers**
   ```bash
   # Server 1: API + 2 GPU workers
   # Server 2: 4 GPU workers only
   # Load balancer in front
   ```

See the project wiki for detailed architecture evolution path.

---

## Project Structure

```
audio-transcription/
├── cmd/server/main.go               # Entry point & server setup
├── internal/
│   ├── handlers/                    # HTTP/WebSocket request handlers
│   │   ├── upload.go                # File upload endpoint
│   │   ├── gdrive.go                # Google Drive download handler
│   │   ├── youtube.go               # YouTube audio extraction
│   │   └── stream.go                # WebSocket streaming handler
│   ├── transcription/               # Audio processing & Whisper integration
│   │   ├── whisper.go               # Python Whisper CLI wrapper
│   │   ├── audio.go                 # FFmpeg audio normalization
│   │   └── diarization.go           # Speaker diarization (future)
│   ├── storage/                     # Persistence layer
│   │   ├── local.go                 # Local filesystem storage
│   │   ├── gdrive_client.go         # Google Drive API client
│   │   └── metadata.go              # SQLite metadata database
│   ├── queue/                       # Concurrent job processing
│   │   ├── worker.go                # Worker pool implementation
│   │   └── jobs.go                  # Job & result types
│   ├── types/                       # Shared type definitions
│   │   └── types.go
│   └── cleanup/                     # Background maintenance
│       └── scheduler.go             # Temp file cleanup scheduler
├── config/config.yaml               # Server & Whisper configuration
├── go.mod
├── go.sum
└── README.md
```

---

## License

MIT License - feel free to use for personal or commercial projects.

---

## Credits

- **[OpenAI Whisper](https://github.com/openai/whisper)** — speech recognition model
- **[Fiber](https://gofiber.io/)** — Express-inspired Go web framework
- **[PyTorch](https://pytorch.org/)** — GPU-accelerated deep learning (CUDA)
- **[yt-dlp](https://github.com/yt-dlp/yt-dlp)** — YouTube audio extraction
- **[Google Drive API](https://developers.google.com/drive)** — cloud storage integration
- **[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)** — pure Go SQLite driver
