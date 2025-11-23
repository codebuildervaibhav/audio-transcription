# Verify Whisper Installation - Run this after setup_whisper.ps1 completes

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Verifying Whisper Installation                  " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Check 1: Python Whisper command
Write-Host "Checking Python Whisper..." -ForegroundColor Yellow
try {
    $whisperVersion = whisper --version 2>&1
    Write-Host "[OK] Whisper command found" -ForegroundColor Green
    
    # Try a simple help command
    $null = whisper --help 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Whisper is working correctly" -ForegroundColor Green
    }
} catch {
    Write-Host "[ERROR] Whisper command not found" -ForegroundColor Red
    Write-Host "Please install: pip install openai-whisper" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# Check 2: Model file
Write-Host "Checking Whisper model file..." -ForegroundColor Yellow
$modelPath = "D:\Development\listner\models\ggml-small.bin"
if (Test-Path $modelPath) {
    $size = (Get-Item $modelPath).Length / 1MB
    $sizeRounded = [math]::Round($size, 2)
    Write-Host "[OK] Model file exists ($sizeRounded MB)" -ForegroundColor Green
} else {
    Write-Host "[WARN] Model file not found at $modelPath" -ForegroundColor Yellow
    Write-Host "      This is OK for Python Whisper - it downloads models automatically" -ForegroundColor Gray
}

Write-Host ""

# Check 3: FFmpeg
Write-Host "Checking FFmpeg..." -ForegroundColor Yellow
try {
    $ffmpegVersion = ffmpeg -version 2>&1 | Select-Object -First 1
    Write-Host "[OK] FFmpeg found" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] FFmpeg not found" -ForegroundColor Red
    Write-Host "Please install: choco install ffmpeg" -ForegroundColor Yellow
}

Write-Host ""

# Check 4: Go
Write-Host "Checking Go..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    Write-Host "[OK] Go found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Go not found" -ForegroundColor Red
    Write-Host "Please install from: https://golang.org/dl/" -ForegroundColor Yellow
}

Write-Host ""

# Check 5: Project structure
Write-Host "Checking project structure..." -ForegroundColor Yellow
$requiredDirs = @(
    "D:\Development\listner\cmd\server",
    "D:\Development\listner\internal\handlers",
    "D:\Development\listner\internal\transcription",
    "D:\Development\listner\models",
    "D:\Development\listner\temp",
    "D:\Development\listner\outputs"
)

$allDirsExist = $true
foreach ($dir in $requiredDirs) {
    if (Test-Path $dir) {
        Write-Host "[OK] $($dir.Replace('D:\Development\listner\', ''))" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Missing: $dir" -ForegroundColor Red
        $allDirsExist = $false
    }
}

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Verification Summary                            " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Final summary
$readyToRun = $true
$whisperOk = $false
$ffmpegOk = $false
$goOk = $false

# Re-check critical components
try { $null = whisper --help 2>&1; if ($LASTEXITCODE -eq 0) { $whisperOk = $true } } catch { }
try { $null = ffmpeg -version 2>&1; if ($LASTEXITCODE -eq 0) { $ffmpegOk = $true } } catch { }
try { $null = go version 2>&1; if ($LASTEXITCODE -eq 0) { $goOk = $true } } catch { }

if ($whisperOk -and $ffmpegOk -and $goOk -and $allDirsExist) {
    Write-Host "[SUCCESS] All dependencies installed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You are ready to start the server:" -ForegroundColor Cyan
    Write-Host "  cd D:\Development\listner" -ForegroundColor White
    Write-Host "  go run cmd\server\main.go" -ForegroundColor White
    Write-Host ""
    Write-Host "The server will start at: http://localhost:3000" -ForegroundColor Yellow
} else {
    Write-Host "[WARNING] Some dependencies are missing:" -ForegroundColor Yellow
    if (-not $whisperOk) { Write-Host "  - Python Whisper (pip install openai-whisper)" -ForegroundColor Red }
    if (-not $ffmpegOk) { Write-Host "  - FFmpeg (choco install ffmpeg)" -ForegroundColor Red }
    if (-not $goOk) { Write-Host "  - Go (https://golang.org/dl/)" -ForegroundColor Red }
    Write-Host ""
    Write-Host "Please install missing dependencies and run this script again" -ForegroundColor Yellow
}

Write-Host ""
