# Quick Whisper Setup for Windows
# Run this script to download the model and optionally install Python Whisper

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Whisper Setup for Audio Transcription Backend  " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Ensure models directory exists
$modelsDir = "D:\Development\listner\models"
if (-not (Test-Path $modelsDir)) {
    New-Item -ItemType Directory -Force -Path $modelsDir | Out-Null
}

# Download Whisper Small model
$modelPath = Join-Path $modelsDir "ggml-small.bin"
$modelUrl = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"

if (Test-Path $modelPath) {
    $size = (Get-Item $modelPath).Length / 1MB
    $sizeRounded = [math]::Round($size, 2)
    Write-Host "[OK] Model already exists ($sizeRounded MB)" -ForegroundColor Green
    Write-Host "     Location: $modelPath" -ForegroundColor Gray
} else {
    Write-Host "Downloading Whisper Small model (~500MB)..." -ForegroundColor Yellow
    Write-Host "This may take 5-10 minutes depending on your connection..." -ForegroundColor Gray
    Write-Host ""
    
    try {
        $ProgressPreference = 'Continue'
        Invoke-WebRequest -Uri $modelUrl -OutFile $modelPath -UseBasicParsing
        
        $size = (Get-Item $modelPath).Length / 1MB
        $sizeRounded = [math]::Round($size, 2)
        Write-Host ""
        Write-Host "[OK] Model downloaded successfully! ($sizeRounded MB)" -ForegroundColor Green
        Write-Host "     Location: $modelPath" -ForegroundColor Gray
    } catch {
        Write-Host "[ERROR] Failed to download model: $_" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Choose Whisper Backend                          " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "You have two options:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Python Whisper (Recommended for Windows)" -ForegroundColor White
Write-Host "   - Easier installation" -ForegroundColor Gray
Write-Host "   - Better Windows compatibility" -ForegroundColor Gray
Write-Host "   - Requires Python 3.8+" -ForegroundColor Gray
Write-Host ""
Write-Host "2. whisper.cpp binary (Advanced)" -ForegroundColor White
Write-Host "   - Requires building with CMake" -ForegroundColor Gray
Write-Host "   - Better performance" -ForegroundColor Gray
Write-Host "   - More complex setup" -ForegroundColor Gray
Write-Host ""

$choice = Read-Host "Enter choice (1 or 2, or 'skip' to decide later)"

if ($choice -eq "1") {
    Write-Host ""
    Write-Host "Installing Python Whisper..." -ForegroundColor Yellow
    
    try {
        $pythonVersion = python --version 2>&1
        Write-Host "[OK] Python detected: $pythonVersion" -ForegroundColor Green
        
        Write-Host "Installing openai-whisper package..." -ForegroundColor Yellow
        pip install -U openai-whisper
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] Python Whisper installed successfully!" -ForegroundColor Green
            Write-Host ""
            Write-Host "Testing installation..." -ForegroundColor Yellow
            $null = whisper --help 2>&1
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[OK] Whisper command verified!" -ForegroundColor Green
                Write-Host ""
                Write-Host "Next step: Update Go code to use Python Whisper" -ForegroundColor Cyan
                Write-Host "See WINDOWS_SETUP.md for code changes" -ForegroundColor Gray
            }
        } else {
            Write-Host "[ERROR] Installation failed" -ForegroundColor Red
        }
    } catch {
        Write-Host "[ERROR] Python not found. Please install Python 3.8+ first:" -ForegroundColor Red
        Write-Host "  https://www.python.org/downloads/" -ForegroundColor Yellow
    }
    
} elseif ($choice -eq "2") {
    Write-Host ""
    Write-Host "Building whisper.cpp requires:" -ForegroundColor Yellow
    Write-Host "1. CMake: choco install cmake" -ForegroundColor Gray
    Write-Host "2. Visual Studio Build Tools or MinGW" -ForegroundColor Gray
    Write-Host ""
    Write-Host "See WINDOWS_SETUP.md for detailed build instructions" -ForegroundColor Cyan
    
} else {
    Write-Host ""
    Write-Host "Setup paused. You can run this script again later." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "  Setup Summary                                   " -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

if (Test-Path $modelPath) {
    Write-Host "[OK] Whisper model ready" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Model not found" -ForegroundColor Red
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Choose and install a Whisper backend (Python or whisper.cpp)" -ForegroundColor White
Write-Host "2. Update Go code if using Python Whisper" -ForegroundColor White
Write-Host "3. Install other dependencies: .\setup.ps1" -ForegroundColor White
Write-Host "4. Start server: go run cmd\server\main.go" -ForegroundColor White
Write-Host ""
Write-Host "For help, see: WINDOWS_SETUP.md" -ForegroundColor Gray
Write-Host ""

