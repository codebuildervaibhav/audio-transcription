# Windows-Specific Notes for Python Whisper Integration

## ✅ What I've Updated

The Go code has been modified to use **Python's OpenAI Whisper** instead of whisper.cpp binary.

### Changes Made:

**File: `internal/transcription/whisper.go`**
- ✅ Calls `whisper` command (Python version)
- ✅ Uses `--output_format json` to get detailed segments
- ✅ Parses Whisper's JSON output format
- ✅ Extracts timestamps for each segment
- ✅ Auto-detects language
- ✅ Works on Windows without building C++ code

## 🎯 Differences: Python Whisper vs whisper.cpp

| Feature | whisper.cpp | Python Whisper |
|---------|-------------|----------------|
| Installation | Build from source (CMake) | `pip install openai-whisper` |
| Model file | Download manually (~500MB) | Auto-downloads on first use |
| Performance | Faster (C++) | Slightly slower (Python) |
| Windows support | Requires build tools | Works out of the box |
| Ease of use | Complex | Simple |

## 📋 What Happens Now

When you start the server, it will:

1. **Look for Python Whisper**
   ```
   whisper --help
   ```
   If not found, error message: "Please install: pip install openai-whisper"

2. **On First Transcription**
   - Python Whisper will download the "small" model (~500MB)
   - This happens automatically
   - Stored in: `~/.cache/whisper/` (user's home directory)

3. **For Each File**
   ```bash
   whisper audio.wav --model small --output_format json --language en
   ```

4. **Returns JSON**
   ```json
   {
     "text": "Full transcript here...",
     "language": "en",
     "segments": [
       {"start": 0.0, "end": 5.2, "text": "First sentence..."},
       {"start": 5.2, "end": 10.1, "text": "Second sentence..."}
     ]
   }
   ```

## 🚀 Next Steps

Once your `setup_whisper.ps1` finishes:

1. **Verify Installation**
   ```powershell
   cd D:\Development\listner
   .\verify_setup.ps1
   ```

2. **Install Go Dependencies**
   ```powershell
   go mod download
   ```

3. **Start the Server**
   ```powershell
   go run cmd\server\main.go
   ```

4. **Test with Postman**
   - Import `Audio_Transcription_API.postman_collection.json`
   - Upload a small audio file
   - Watch the magic happen! ✨

## ⚠️ Important Notes

### Model Storage
- The `ggml-small.bin` file you downloaded is **NOT used by Python Whisper**
- Python Whisper stores models in: `C:\Users\<YourName>\.cache\whisper\`
- First run will download ~500MB to cache
- You can delete `D:\Development\listner\models\ggml-small.bin` if desired

### Performance Expectations
- **First transcription**: Slower (downloading model)
- **Subsequent**: ~30-60 seconds per minute of audio
- Python Whisper uses CPU by default
- For GPU acceleration: Install PyTorch with CUDA

### Common Commands

**Test Whisper manually:**
```powershell
# Create a test audio file or use existing
whisper test_audio.mp3 --model small --language en
```

**Check Whisper models:**
```powershell
# Models stored here:
dir $env:USERPROFILE\.cache\whisper
```

**Upgrade Whisper:**
```powershell
pip install --upgrade openai-whisper
```

## 🔧 Troubleshooting

### Issue: "whisper command not found"
**Solution:**
```powershell
pip install openai-whisper
# Or if pip points to wrong Python:
python -m pip install openai-whisper
```

### Issue: "No module named 'whisper'"
**Solution:** Ensure you're using the same Python that pip installed to:
```powershell
python -m pip install openai-whisper
python -c "import whisper; print(whisper.__version__)"
```

### Issue: Very slow transcription
**Solution:** Python Whisper is CPU-bound. For faster processing:
1. Use smaller model: `tiny` or `base`
2. Install GPU version of PyTorch (if you have NVIDIA GPU)
3. Or switch back to whisper.cpp (faster C++ implementation)

### Issue: Out of memory
**Solution:**
- Use smaller model: Change config to use `tiny` or `base`
- Or reduce worker count in `config/config.yaml` (default: 4)

## 🎓 How It Works

```
User uploads file
      ↓
Go saves to /temp/
      ↓
Go calls: whisper temp/file.wav --model small --output_format json
      ↓
Python Whisper processes (30-60s per minute of audio)
      ↓
Returns JSON with text + timestamps
      ↓
Go parses JSON
      ↓
Saves to /outputs/ and Google Drive
      ↓
Returns job completion
```

## 📊 Expected Output

After running the server and uploading a file, you'll see logs like:

```
Transcribing with Python Whisper: temp/abc123.wav
Whisper output: Detecting language using up to the first 30 seconds...
Processing audio...
Whisper output: 100%|██████████| 100/100 [00:45<00:00,  2.21it/s]
Transcription completed: 42 segments, 125.50s duration
```

## ✅ Ready to Test!

Once your installation completes, run:
```powershell
.\verify_setup.ps1
```

If all checks pass, start the server:
```powershell
go run cmd\server\main.go
```

Then test with Postman or curl! 🚀
