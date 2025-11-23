# Quick Setup Script for Windows (PowerShell)
# Run this after installing Go, FFmpeg, and downloading Whisper model

Write-Host "🚀 Audio Transcription Backend - Quick Setup" -ForegroundColor Green
Write-Host ""

# Check Go installation
Write-Host "Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✓ Go installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go not found. Please install Go 1.21+ from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

# Check FFmpeg
Write-Host ""
Write-Host "Checking FFmpeg..." -ForegroundColor Yellow
try {
    $ffmpegVersion = ffmpeg -version | Select-Object -First 1
    Write-Host "✓ FFmpeg installed" -ForegroundColor Green
} catch {
    Write-Host "✗ FFmpeg not found. Install with: choco install ffmpeg" -ForegroundColor Red
    exit 1
}

# Check Whisper model
Write-Host ""
Write-Host "Checking Whisper model..." -ForegroundColor Yellow
if (Test-Path "./models/ggml-small.bin") {
    $modelSize = (Get-Item "./models/ggml-small.bin").Length / 1MB
    Write-Host "✓ Whisper model found (${modelSize}MB)" -ForegroundColor Green
} else {
    Write-Host "✗ Whisper model not found at ./models/ggml-small.bin" -ForegroundColor Red
    Write-Host "  Download instructions:" -ForegroundColor Yellow
    Write-Host "  1. git clone https://github.com/ggerganov/whisper.cpp" -ForegroundColor Yellow
    Write-Host "  2. cd whisper.cpp" -ForegroundColor Yellow
    Write-Host "  3. bash ./models/download-ggml-model.sh small" -ForegroundColor Yellow
    Write-Host "  4. Copy models/ggml-small.bin to this project's models/ folder" -ForegroundColor Yellow
    exit 1
}

# Check Whisper binary
Write-Host ""
Write-Host "Checking Whisper binary..." -ForegroundColor Yellow
if (Test-Path "./whisper") {
    Write-Host "✓ Whisper binary found" -ForegroundColor Green
} else {
    Write-Host "⚠ Whisper binary not found (./whisper)" -ForegroundColor Yellow
    Write-Host "  This is required for transcription. Build whisper.cpp and copy the binary here." -ForegroundColor Yellow
}

# Install Go dependencies
Write-Host ""
Write-Host "Installing Go dependencies..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to install dependencies" -ForegroundColor Red
    exit 1
}

# Optional: Check yt-dlp for YouTube support
Write-Host ""
Write-Host "Checking yt-dlp (optional, for YouTube)..." -ForegroundColor Yellow
try {
    $ytdlpVersion = yt-dlp --version
    Write-Host "✓ yt-dlp installed: $ytdlpVersion" -ForegroundColor Green
} catch {
    Write-Host "⚠ yt-dlp not found (YouTube feature will not work)" -ForegroundColor Yellow
    Write-Host "  Install with: pip install yt-dlp" -ForegroundColor Yellow
}

# Check Google Drive credentials (optional)
Write-Host ""
Write-Host "Checking Google Drive credentials (optional)..." -ForegroundColor Yellow
if (Test-Path "./credentials.json") {
    Write-Host "✓ credentials.json found" -ForegroundColor Green
} else {
    Write-Host "⚠ credentials.json not found (Google Drive backup disabled)" -ForegroundColor Yellow
    Write-Host "  Transcripts will be saved locally only" -ForegroundColor Yellow
    Write-Host "  To enable: Place Google Drive OAuth credentials in credentials.json" -ForegroundColor Yellow
}

# Done
Write-Host ""
Write-Host "================================================" -ForegroundColor Green
Write-Host "✓ Setup complete!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host ""
Write-Host "To start the server:" -ForegroundColor Cyan
Write-Host "  go run cmd/server/main.go" -ForegroundColor White
Write-Host ""
Write-Host "Or build and run:" -ForegroundColor Cyan
Write-Host "  go build -o transcription-server.exe cmd/server/main.go" -ForegroundColor White
Write-Host "  .\transcription-server.exe" -ForegroundColor White
Write-Host ""
Write-Host "Server will start at: http://localhost:3000" -ForegroundColor Yellow
Write-Host "API Docs: README.md" -ForegroundColor Yellow
Write-Host ""
