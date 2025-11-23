#!/bin/bash
# Test Script for Audio Transcription Backend
# This script tests all endpoints to verify the system is working

BASE_URL="http://localhost:3000"

echo "🧪 Audio Transcription Backend - Integration Tests"
echo "=================================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

if [ "$status_code" -eq 200 ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "  Response: $body"
else
    echo -e "${RED}✗ Health check failed (Status: $status_code)${NC}"
    exit 1
fi

echo ""

# Test 2: List Transcripts (should be empty initially)
echo -e "${YELLOW}Test 2: List Transcripts${NC}"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/transcripts")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

if [ "$status_code" -eq 200 ]; then
    echo -e "${GREEN}✓ Transcripts endpoint working${NC}"
    echo "  Response: $body"
else
    echo -e "${RED}✗ Transcripts endpoint failed (Status: $status_code)${NC}"
fi

echo ""

# Test 3: Upload (requires a test file)
echo -e "${YELLOW}Test 3: File Upload${NC}"
if [ -f "test_audio.mp3" ]; then
    response=$(curl -s -w "\n%{http_code}" -F "file=@test_audio.mp3" -F "name=IntegrationTest" "$BASE_URL/upload")
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" -eq 200 ]; then
        job_id=$(echo "$body" | grep -o '"job_id":"[^"]*' | cut -d'"' -f4)
        echo -e "${GREEN}✓ File upload successful${NC}"
        echo "  Job ID: $job_id"
        echo "  Full response: $body"
    else
        echo -e "${RED}✗ File upload failed (Status: $status_code)${NC}"
        echo "  Response: $body"
    fi
else
    echo -e "${YELLOW}⚠ Skipped - test_audio.mp3 not found${NC}"
    echo "  Create a small MP3 file named 'test_audio.mp3' to test uploads"
fi

echo ""

# Test 4: Google Drive (requires a public share link)
echo -e "${YELLOW}Test 4: Google Drive Link${NC}"
if [ -n "$TEST_GDRIVE_URL" ]; then
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/gdrive" \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"$TEST_GDRIVE_URL\",\"name\":\"GDriveTest\"}")
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" -eq 200 ]; then
        echo -e "${GREEN}✓ Google Drive processing started${NC}"
        echo "  Response: $body"
    else
        echo -e "${RED}✗ Google Drive processing failed (Status: $status_code)${NC}"
        echo "  Response: $body"
    fi
else
    echo -e "${YELLOW}⚠ Skipped - Set TEST_GDRIVE_URL environment variable to test${NC}"
    echo "  Example: export TEST_GDRIVE_URL='https://drive.google.com/file/d/.../view'"
fi

echo ""

# Test 5: YouTube (requires yt-dlp)
echo -e "${YELLOW}Test 5: YouTube Capture${NC}"
if command -v yt-dlp &> /dev/null; then
    if [ -n "$TEST_YOUTUBE_URL" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/youtube" \
            -H "Content-Type: application/json" \
            -d "{\"url\":\"$TEST_YOUTUBE_URL\",\"name\":\"YouTubeTest\"}")
        status_code=$(echo "$response" | tail -n 1)
        body=$(echo "$response" | head -n -1)
        
        if [ "$status_code" -eq 200 ]; then
            echo -e "${GREEN}✓ YouTube capture started${NC}"
            echo "  Response: $body"
        else
            echo -e "${RED}✗ YouTube capture failed (Status: $status_code)${NC}"
            echo "  Response: $body"
        fi
    else
        echo -e "${YELLOW}⚠ Skipped - Set TEST_YOUTUBE_URL environment variable to test${NC}"
        echo "  Example: export TEST_YOUTUBE_URL='https://www.youtube.com/watch?v=...'"
    fi
else
    echo -e "${YELLOW}⚠ Skipped - yt-dlp not installed${NC}"
    echo "  Install with: pip install yt-dlp"
fi

echo ""

# Summary
echo "=================================================="
echo -e "${GREEN}Testing complete!${NC}"
echo ""
echo "Check the outputs/ directory for transcription results"
echo "Monitor worker logs in the server console"
echo ""
