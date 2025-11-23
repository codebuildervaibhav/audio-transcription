# Windows Setup Guide for Whisper.cpp

## Issue: `make` not available on Windows

You have 3 options:

## ✅ **Option 1: Use CMake (Recommended)**

### Step 1: Install CMake
```powershell
# Using Chocolatey
choco install cmake

# Or download from: https://cmake.org/download/
```

### Step 2: Build with CMake
```powershell
cd C:\Users\vaibh\whisper.cpp

# Create build directory
mkdir build
cd build

# Configure
cmake ..

# Build (this will take a few minutes)
cmake --build . --config Release

# The executable will be in: build\bin\Release\main.exe
```

### Step 3: Copy files to your project
```powershell
# Copy the model downloader script
cd C:\Users\vaibh\whisper.cpp

# Download model using PowerShell instead of bash
Invoke-WebRequest -Uri "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin" -OutFile "models\ggml-small.bin"

# Copy to your project (this will be ~500MB)
Copy-Item models\ggml-small.bin D:\Development\listner\models\

# Copy the whisper executable
Copy-Item build\bin\Release\main.exe D:\Development\listner\whisper.exe
```

---

## ✅ **Option 2: Direct Model Download (Fastest)**

Skip the build and just download what you need:

```powershell
# Create models directory if it doesn't exist
New-Item -ItemType Directory -Force -Path "D:\Development\listner\models"

# Download the Whisper Small model (~500MB)
$url = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"
$output = "D:\Development\listner\models\ggml-small.bin"

Write-Host "Downloading Whisper model (~500MB, this may take a few minutes)..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $url -OutFile $output -UseBasicParsing

Write-Host "✓ Model downloaded successfully!" -ForegroundColor Green
Write-Host "Location: $output"

# Download pre-built Windows binary
$binaryUrl = "https://github.com/ggerganov/whisper.cpp/releases/download/v1.5.4/whisper-bin-x64.zip"
$zipPath = "$env:TEMP\whisper-bin.zip"

Write-Host "Downloading Whisper binary..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $binaryUrl -OutFile $zipPath -UseBasicParsing

# Extract
Expand-Archive -Path $zipPath -DestinationPath "D:\Development\listner\" -Force

Write-Host "✓ Setup complete!" -ForegroundColor Green
```

**Save this as `download_whisper.ps1` and run:**
```powershell
cd D:\Development\listner
.\download_whisper.ps1
```

---

## ✅ **Option 3: Use Visual Studio (For developers)**

If you have Visual Studio installed:

```powershell
cd C:\Users\vaibh\whisper.cpp

# Open the solution file
start whisper.sln

# In Visual Studio:
# 1. Select "Release" configuration
# 2. Build → Build Solution
# 3. Find main.exe in x64\Release\
```

---

## 🎯 **Simplest Solution - Just Get the Model**

If you just want to test without building:

### Download the model manually:
1. Go to: https://huggingface.co/ggerganov/whisper.cpp/tree/main
2. Download `ggml-small.bin` (~500MB)
3. Save to: `D:\Development\listner\models\ggml-small.bin`

### Use Python whisper instead:
```powershell
# Install Python Whisper (easier alternative)
pip install openai-whisper

# Test it
whisper test_audio.mp3 --model small --output_dir ./outputs
```

Then modify the Go code to call Python's whisper instead of whisper.cpp.

---

## 📝 **Updated Go Code for Python Whisper**

If you go the Python route, update `internal/transcription/whisper.go`:

```go
// Transcribe using Python's whisper
cmd := exec.Command("whisper",
    audioPath,
    "--model", "small",
    "--output_dir", "temp",
    "--output_format", "txt",
)

output, err := cmd.CombinedOutput()
```

---

## ⚡ **Quick Test Command**

After downloading the model, test if everything works:

```powershell
# If using CMake build:
.\build\bin\Release\main.exe -m models\ggml-small.bin -f test_audio.wav

# If using Python whisper:
whisper test_audio.mp3 --model small
```

---

## 🔧 **Fix the Line Ending Issue**

The `$'\r'` errors are because the bash scripts have Windows line endings.

**Fix:**
```powershell
# Install dos2unix (converts line endings)
choco install dos2unix

# Convert the script
dos2unix C:\Users\vaibh\whisper.cpp\models\download-ggml-model.sh

# Now run it
bash ./models/download-ggml-model.sh small
```

---

## 📦 **What I Recommend for You**

**Best approach for Windows:**

1. **Skip building whisper.cpp** - use the direct download
2. **Use Option 2** - run the PowerShell script below
3. **Start your server** and test

**Run this now:**

```powershell
cd D:\Development\listner

# Download model
$url = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"
Invoke-WebRequest -Uri $url -OutFile "models\ggml-small.bin"

Write-Host "✓ Model ready! Size:" (Get-Item models\ggml-small.bin).Length / 1MB "MB"
```

Then we can either:
- Install Python Whisper (easier)
- Or build whisper.cpp with CMake (more complex)

**Which would you prefer?**
