# Using the Postman Collection

## Import Instructions

1. **Open Postman**
2. Click `Import` button (top left)
3. Select `Audio_Transcription_API.postman_collection.json`
4. Collection will appear in your sidebar

## Configuration

### Environment Variables
The collection uses these variables (auto-configured):
- `base_url`: `http://localhost:3000` (change if using different port)
- `job_id`: Auto-populated after requests

### Change Base URL
If your server runs on a different port:
1. Click on the collection name
2. Go to "Variables" tab
3. Change `base_url` value

## Testing Order

### Quick Test Flow (5 minutes)

1. **Health Check** ✓
   - Verify server is running
   - Should return: `{"status":"healthy"}`

2. **Upload Audio File** ✓
   - Click on "file" field in Body
   - Select an audio file (MP3, WAV, etc.)
   - Send request
   - **Save the `job_id` from response**

3. **List Transcripts**
   - Wait 30-60 seconds for processing
   - Check if your file appears in the list

4. **Check Output Folder**
   - Navigate to `D:\Development\listner\outputs\YYYY\MM\DD\`
   - Find your transcript `.txt` file

### Advanced Tests

#### Google Drive Test
1. Upload an audio file to Google Drive
2. Share with "Anyone with the link"
3. Copy the share URL
4. Open "Process Google Drive Link" request
5. Replace `YOUR_FILE_ID_HERE` with actual URL
6. Send request

#### YouTube Test
**Requirements:** yt-dlp installed (`pip install yt-dlp`)

1. Copy any YouTube URL
2. Open "Capture YouTube Audio" request
3. Replace URL in request body
4. Send request
5. **Note:** May take 1-5 minutes for download + processing

## Automated Tests

Each request includes automated tests that verify:
- ✓ Status code is 200
- ✓ Response contains expected fields
- ✓ Job IDs are properly formatted
- ✓ Data types are correct

**View test results:**
- Send a request
- Look at the "Test Results" tab at the bottom

## Common Issues

### Issue: "file" field shows "No file selected"
**Solution:** 
1. Click on the "file" key in Body → form-data
2. Hover over the right side of the "file" row
3. Click "Select Files" button
4. Choose your audio file

### Issue: 404 Not Found
**Solution:**
- Ensure server is running (`go run cmd/server/main.go`)
- Check `base_url` variable is correct
- Verify port 3000 is not blocked

### Issue: Google Drive returns "File not accessible"
**Solution:**
- Ensure file is shared as "Anyone with the link can view"
- Check URL is in correct format
- Try with a smaller file first (< 50MB)

### Issue: YouTube capture fails
**Solution:**
- Verify yt-dlp is installed: `yt-dlp --version`
- Try with a short video first (< 5 minutes)
- Check video is not age-restricted or region-locked

## WebSocket Testing

Postman Desktop App supports WebSocket:
1. Create new WebSocket Request
2. URL: `ws://localhost:3000/ws/stream`
3. Connect
4. Send text message: `TestRecording` (sets name)
5. Send binary data (audio chunks)
6. Send text message: `END`
7. Receive JSON with `job_id`

**Alternative:** Use browser-based tools like [WebSocket King](https://websocketking.com/)

## Request Examples

### Upload with curl (alternative)
```bash
curl -F "file=@audio.mp3" -F "name=TestAudio" http://localhost:3000/upload
```

### Google Drive with curl
```bash
curl -X POST http://localhost:3000/gdrive \
  -H "Content-Type: application/json" \
  -d '{"url":"https://drive.google.com/file/d/.../view","name":"Test"}'
```

### YouTube with curl
```bash
curl -X POST http://localhost:3000/youtube \
  -H "Content-Type: application/json" \
  -d '{"url":"https://youtube.com/watch?v=...","name":"Video"}'
```

## Response Fields Explained

### Job Response
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",  // Unique job identifier
  "status": "queued",                                 // QUEUED | PROCESSING | COMPLETED | FAILED
  "message": "File uploaded successfully..."          // Human-readable message
}
```

### Transcript Metadata
```json
{
  "job_id": "...",
  "request_name": "MyAudio",           // Name you provided
  "source_type": "upload",             // upload | gdrive | youtube | stream
  "gdrive_url": "https://...",         // Google Drive link (if enabled)
  "local_path": "./outputs/.../...",   // Local file path
  "created_at": "2025-11-23T...",      // ISO 8601 timestamp
  "duration": 125.5,                   // Audio duration in seconds
  "word_count": 342                    // Number of words in transcript
}
```

## Tips

1. **Start Small**
   - Test with short audio files (< 1 minute) first
   - Verify the full pipeline works before long files

2. **Monitor Server Logs**
   - Watch the terminal where server is running
   - You'll see real-time processing updates

3. **Check Output Files**
   - Look in `outputs/YYYY/MM/DD/` for results
   - `.txt` = transcript
   - `_meta.json` = metadata with timestamps

4. **Use Environment Variables**
   - Store frequently used URLs in Postman environments
   - Switch between dev/staging/prod easily

## Next Steps

After testing with Postman:
1. Build a frontend (see QUICKSTART.md for examples)
2. Integrate into your application
3. Deploy to production (see implementation_plan.md for scaling)

---

Need help? Check the main [README.md](./README.md) for troubleshooting!
