# Quick Start Guide

## 🚀 Get Running in 5 Minutes

### Step 1: Install Prerequisites (15 minutes first time)

#### Windows
```powershell
# 1. Install Go (if not installed)
# Download from: https://golang.org/dl/
# Run installer, then verify:
go version

# 2. Install FFmpeg
choco install ffmpeg
# Or download from: https://ffmpeg.org/download.html

# 3. Install Python (for yt-dlp - optional)
choco install python
pip install yt-dlp
```

#### macOS/Linux
```bash
# Go
brew install go  # macOS
# or download from https://golang.org/dl/

# FFmpeg
brew install ffmpeg  # macOS
sudo apt install ffmpeg  # Ubuntu/Debian

# yt-dlp (optional)
pip install yt-dlp
```

### Step 2: Get Whisper Model (5 minutes)

```bash
# Clone whisper.cpp
git clone https://github.com/ggerganov/whisper.cpp
cd whisper.cpp

# Build (Windows: use make.exe or Visual Studio)
make

# Download model (~500MB, one-time download)
bash ./models/download-ggml-model.sh small

# Copy model to your project
cp models/ggml-small.bin D:/Development/listner/models/

# Copy whisper binary
cp main D:/Development/listner/whisper
# On Windows, this might be main.exe or whisper.exe
```

### Step 3: Setup Project (1 minute)

```powershell
cd D:\Development\listner

# Run automated setup
.\setup.ps1

# This will check:
# ✓ Go installation
# ✓ FFmpeg
# ✓ Whisper model
# ✓ Install Go dependencies
```

### Step 4: Run Server (immediate)

```bash
# Start server
go run cmd/server/main.go

# You should see:
# 🚀 Server starting on 0.0.0.0:3000
# 📝 Endpoints: ...
```

Server is now running at **http://localhost:3000**!

### Step 5: Test It (30 seconds)

#### Quick Test - Health Check
```bash
curl http://localhost:3000/health
# Expected: {"status":"healthy","version":"1.0.0"}
```

#### Full Test - Upload a File
1. Get a small audio file (MP3, WAV, etc.)
2. Upload:
```bash
curl -F "file=@your_audio.mp3" -F "name=TestRecording" http://localhost:3000/upload
```
3. You'll get a job ID - processing happens in background
4. Check `outputs/` directory for the transcript!

---

## 🎯 Common Use Cases

### Use Case 1: Transcribe a Local File
```bash
curl -F "file=@meeting.mp3" -F "name=TeamMeeting" http://localhost:3000/upload
```

### Use Case 2: Transcribe from Google Drive
```bash
curl -X POST http://localhost:3000/gdrive \
  -H "Content-Type: application/json" \
  -d '{"url":"https://drive.google.com/file/d/YOUR_FILE_ID/view","name":"Presentation"}'
```

### Use Case 3: Record and Transcribe Live Audio

**Frontend (HTML + JavaScript):**
```html
<!DOCTYPE html>
<html>
<body>
  <button onclick="startRecording()">Start Recording</button>
  <button onclick="stopRecording()">Stop Recording</button>
  <div id="status"></div>
  <script>
    let mediaRecorder;
    let ws;

    async function startRecording() {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      mediaRecorder = new MediaRecorder(stream);
      
      ws = new WebSocket('ws://localhost:3000/ws/stream');
      
      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0 && ws.readyState === WebSocket.OPEN) {
          ws.send(event.data);
        }
      };
      
      mediaRecorder.start(1000); // Send chunks every 1 second
      document.getElementById('status').innerText = '🔴 Recording...';
    }

    function stopRecording() {
      mediaRecorder.stop();
      ws.send('END');
      
      ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        document.getElementById('status').innerText = 
          `✅ Transcription started! Job ID: ${response.job_id}`;
      };
    }
  </script>
</body>
</html>
```

### Use Case 4: Transcribe a YouTube Video
```bash
curl -X POST http://localhost:3000/youtube \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.youtube.com/watch?v=VIDEO_ID","name":"TED_Talk"}'
```

---

## 📂 Where Are My Transcripts?

### Local Storage
```
outputs/
└── 2025/
    └── 11/
        └── 23/
            ├── 20251123_211514_TeamMeeting.txt        # ← Your transcript
            └── 20251123_211514_TeamMeeting_meta.json  # ← Metadata
```

### Google Drive (if configured)
```
Transcripts/  (in your Google Drive)
└── 2025/
    └── 11/
        └── 23/
            ├── 20251123_211514_TeamMeeting.txt
            └── 20251123_211514_TeamMeeting_meta.json
```

---

## ⚙️ Optional: Enable Google Drive Backup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project
3. Enable "Google Drive API"
4. Create OAuth 2.0 Client ID (Desktop app)
5. Download `credentials.json`
6. Place in project root: `D:\Development\listner\credentials.json`
7. Restart server - it will open a browser for authentication
8. Done! Transcripts now auto-backup to Google Drive

---

## 🐛 Troubleshooting

### "Whisper model not found"
```bash
# Make sure the model is in the right place:
ls D:\Development\listner\models\ggml-small.bin

# If not, download it:
cd whisper.cpp
bash ./models/download-ggml-model.sh small
cp models/ggml-small.bin D:/Development/listner/models/
```

### "ffmpeg: command not found"
```bash
# Windows: Install via Chocolatey
choco install ffmpeg

# Or download from: https://ffmpeg.org/download.html
# Add to PATH
```

### "Port 3000 already in use"
Edit `config/config.yaml`:
```yaml
server:
  port: 8080  # Use a different port
```

### Server crashes during transcription
- **Likely cause:** Whisper binary not found or incompatible
- **Fix:** Rebuild whisper.cpp for your architecture
```bash
cd whisper.cpp
make clean
make
cp main D:/Development/listner/whisper
```

---

## 📖 Learn More

- **Full Documentation:** [README.md](./README.md)
- **API Reference:** [README.md#api-usage](./README.md#api-usage)
- **Architecture:** [project_context.md](./project_context.md)
- **Scaling Guide:** [implementation_plan.md](./implementation_plan.md)

---

## 🎉 You're All Set!

Your audio transcription backend is now running. Try uploading a file and watch the magic happen! 🚀

**Questions?** Check the [Troubleshooting](#troubleshooting) section above.
