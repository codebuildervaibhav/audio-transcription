package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/vaibh/audio-transcription/internal/types"
)

// DriveClient handles uploading to Google Drive
type DriveClient struct {
	service    *drive.Service
	folderName string
	folderID   string
}

// NewDriveClient creates a new Google Drive client
func NewDriveClient(credentialsFile, tokenFile, folderName string) (*DriveClient, error) {
	ctx := context.Background()

	// Read credentials
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %v", err)
	}

	client := getClient(config, tokenFile)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %v", err)
	}

	dc := &DriveClient{
		service:    srv,
		folderName: folderName,
	}

	// Find or create the root folder
	if err := dc.ensureFolder(); err != nil {
		return nil, err
	}

	return dc, nil
}

// getClient retrieves a token, saves the token, then returns the generated client
func getClient(config *oauth2.Config, tokenFile string) *http.Client {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb requests a token from the web
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)
	fmt.Print("Enter authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		panic(fmt.Sprintf("Unable to read authorization code: %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		panic(fmt.Sprintf("Unable to retrieve token from web: %v", err))
	}
	return tok
}

// tokenFromFile retrieves a token from a local file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Sprintf("Unable to cache oauth token: %v", err))
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// ensureFolder finds or creates the root folder
func (dc *DriveClient) ensureFolder() error {
	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and trashed=false", 
		dc.folderName)
	
	r, err := dc.service.Files.List().Q(query).Spaces("drive").Fields("files(id, name)").Do()
	if err != nil {
		return fmt.Errorf("unable to search for folder: %v", err)
	}

	if len(r.Files) > 0 {
		dc.folderID = r.Files[0].Id
		return nil
	}

	// Create folder
	folder := &drive.File{
		Name:     dc.folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	file, err := dc.service.Files.Create(folder).Fields("id").Do()
	if err != nil {
		return fmt.Errorf("unable to create folder: %v", err)
	}

	dc.folderID = file.Id
	return nil
}

// Upload uploads transcript and metadata to Google Drive
func (dc *DriveClient) Upload(requestName string, result *types.TranscriptionResult) (string, error) {
	// Create dated folder structure: Transcripts/2025/01/23/
	now := time.Now()
	folderID, err := dc.ensureDateFolder(now)
	if err != nil {
		return "", err
	}

	// Generate filename
	timestamp := now.Format("20060102_150405")
	baseFilename := fmt.Sprintf("%s_%s", timestamp, sanitizeFilename(requestName))

	// Upload transcript text
	txtFile := &drive.File{
		Name:    baseFilename + ".txt",
		Parents: []string{folderID},
	}

	_, err = dc.service.Files.Create(txtFile).Media(
		createReaderFromString(result.Text)).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload transcript: %v", err)
	}

	// Upload metadata JSON
	metadata := map[string]interface{}{
		"job_id":           result.JobID,
		"request_name":     requestName,
		"duration_seconds": result.Duration,
		"word_count":       result.WordCount,
		"model_used":       "whisper-small",
		"language":         result.Language,
		"created_at":       result.ProcessedAt,
		"segments":         result.Segments,
	}

	metaJSON, _ := json.MarshalIndent(metadata, "", "  ")
	
	metaFile := &drive.File{
		Name:    baseFilename + "_meta.json",
		Parents: []string{folderID},
	}

	createdMeta, err := dc.service.Files.Create(metaFile).Media(
		createReaderFromBytes(metaJSON)).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload metadata: %v", err)
	}

	// Return shareable link
	fileURL := fmt.Sprintf("https://drive.google.com/file/d/%s/view", createdMeta.Id)
	return fileURL, nil
}

// ensureDateFolder creates nested year/month/day folders
func (dc *DriveClient) ensureDateFolder(t time.Time) (string, error) {
	// Create year folder
	yearID, err := dc.findOrCreateFolder(fmt.Sprintf("%d", t.Year()), dc.folderID)
	if err != nil {
		return "", err
	}

	// Create month folder
	monthID, err := dc.findOrCreateFolder(fmt.Sprintf("%02d", t.Month()), yearID)
	if err != nil {
		return "", err
	}

	// Create day folder
	dayID, err := dc.findOrCreateFolder(fmt.Sprintf("%02d", t.Day()), monthID)
	if err != nil {
		return "", err
	}

	return dayID, nil
}

// findOrCreateFolder finds or creates a folder with the given parent
func (dc *DriveClient) findOrCreateFolder(name, parentID string) (string, error) {
	query := fmt.Sprintf("name='%s' and '%s' in parents and mimeType='application/vnd.google-apps.folder' and trashed=false",
		name, parentID)

	r, err := dc.service.Files.List().Q(query).Spaces("drive").Fields("files(id)").Do()
	if err != nil {
		return "", err
	}

	if len(r.Files) > 0 {
		return r.Files[0].Id, nil
	}

	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	file, err := dc.service.Files.Create(folder).Fields("id").Do()
	if err != nil {
		return "", err
	}

	return file.Id, nil
}

// Helper to create reader from string
func createReaderFromString(s string) *os.File {
	tmpFile, _ := os.CreateTemp("", "upload-*.txt")
	tmpFile.WriteString(s)
	tmpFile.Seek(0, 0)
	return tmpFile
}

// Helper to create reader from bytes
func createReaderFromBytes(b []byte) *os.File {
	tmpFile, _ := os.CreateTemp("", "upload-*.json")
	tmpFile.Write(b)
	tmpFile.Seek(0, 0)
	return tmpFile
}
