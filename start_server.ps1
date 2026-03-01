# Audio Transcription Service — Quick Start
# Ensures PATH includes Go and Python, then launches the API server.

$env:GOROOT = "D:\go"
$env:Path += ";C:\Users\vaibh\AppData\Local\Programs\Python\Python313\Scripts"

Write-Host ""
Write-Host "  Audio Transcription Service" -ForegroundColor Cyan
Write-Host "  Starting on http://localhost:3000 ..." -ForegroundColor DarkGray
Write-Host ""

go run ./cmd/server
