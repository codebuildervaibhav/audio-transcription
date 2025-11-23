package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// MetadataDB handles SQLite database operations
type MetadataDB struct {
	db *sql.DB
}

// NewMetadataDB creates a new metadata database
func NewMetadataDB(dbPath string) (*MetadataDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS transcripts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT NOT NULL UNIQUE,
		request_name TEXT NOT NULL,
		source_type TEXT NOT NULL,
		gdrive_url TEXT,
		local_path TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		duration REAL,
		word_count INTEGER
	);

	CREATE INDEX IF NOT EXISTS idx_created_at ON transcripts(created_at);
	CREATE INDEX IF NOT EXISTS idx_request_name ON transcripts(request_name);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	return &MetadataDB{db: db}, nil
}

// SaveTranscript saves transcript metadata to the database
func (mdb *MetadataDB) SaveTranscript(
	jobID, requestName, sourceType, gdriveURL, localPath string,
	duration float64, wordCount int,
) error {
	query := `
	INSERT INTO transcripts (job_id, request_name, source_type, gdrive_url, local_path, created_at, duration, word_count)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := mdb.db.Exec(query, jobID, requestName, sourceType, gdriveURL, localPath, 
		time.Now(), duration, wordCount)
	if err != nil {
		return fmt.Errorf("failed to save transcript metadata: %v", err)
	}

	return nil
}

// GetTranscript retrieves transcript metadata by job ID
func (mdb *MetadataDB) GetTranscript(jobID string) (map[string]interface{}, error) {
	query := `
	SELECT job_id, request_name, source_type, gdrive_url, local_path, created_at, duration, word_count
	FROM transcripts WHERE job_id = ?
	`

	row := mdb.db.QueryRow(query, jobID)

	var (
		jid, name, source, gdrive, local string
		createdAt                         time.Time
		duration                          float64
		wordCount                         int
	)

	err := row.Scan(&jid, &name, &source, &gdrive, &local, &createdAt, &duration, &wordCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %v", err)
	}

	return map[string]interface{}{
		"job_id":       jid,
		"request_name": name,
		"source_type":  source,
		"gdrive_url":   gdrive,
		"local_path":   local,
		"created_at":   createdAt,
		"duration":     duration,
		"word_count":   wordCount,
	}, nil
}

// ListTranscripts returns all transcripts
func (mdb *MetadataDB) ListTranscripts(limit int) ([]map[string]interface{}, error) {
	query := `
	SELECT job_id, request_name, source_type, gdrive_url, local_path, created_at, duration, word_count
	FROM transcripts ORDER BY created_at DESC LIMIT ?
	`

	rows, err := mdb.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list transcripts: %v", err)
	}
	defer rows.Close()

	var transcripts []map[string]interface{}

	for rows.Next() {
		var (
			jid, name, source, gdrive, local string
			createdAt                         time.Time
			duration                          float64
			wordCount                         int
		)

		if err := rows.Scan(&jid, &name, &source, &gdrive, &local, &createdAt, &duration, &wordCount); err != nil {
			continue
		}

		transcripts = append(transcripts, map[string]interface{}{
			"job_id":       jid,
			"request_name": name,
			"source_type":  source,
			"gdrive_url":   gdrive,
			"local_path":   local,
			"created_at":   createdAt,
			"duration":     duration,
			"word_count":   wordCount,
		})
	}

	return transcripts, nil
}

// Close closes the database connection
func (mdb *MetadataDB) Close() error {
	return mdb.db.Close()
}
