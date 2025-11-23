package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"gopkg.in/yaml.v3"

	"github.com/vaibh/audio-transcription/internal/cleanup"
	"github.com/vaibh/audio-transcription/internal/handlers"
	"github.com/vaibh/audio-transcription/internal/queue"
	"github.com/vaibh/audio-transcription/internal/storage"
	"github.com/vaibh/audio-transcription/internal/transcription"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	
	Whisper struct {
		Model     string `yaml:"model"`
		ModelPath string `yaml:"model_path"`
		Threads   int    `yaml:"threads"`
	} `yaml:"whisper"`
	
	Workers struct {
		Count int `yaml:"count"`
	} `yaml:"workers"`
	
	Storage struct {
		TempDir   string `yaml:"temp_dir"`
		OutputDir string `yaml:"output_dir"`
		Database  string `yaml:"database"`
	} `yaml:"storage"`
	
	Cleanup struct {
		IntervalMinutes int `yaml:"interval_minutes"`
		MaxAgeHours     int `yaml:"max_age_hours"`
	} `yaml:"cleanup"`
	
	GoogleDrive struct {
		CredentialsFile string `yaml:"credentials_file"`
		TokenFile       string `yaml:"token_file"`
		FolderName      string `yaml:"folder_name"`
	} `yaml:"google_drive"`
	
	Limits struct {
		MaxFileSizeMB     int `yaml:"max_file_size_mb"`
		MaxDurationMinutes int `yaml:"max_duration_minutes"`
	} `yaml:"limits"`
}

func main() {
	// Load configuration
	config, err := loadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Ensure directories exist
	if err := cleanup.EnsureTempDirExists(config.Storage.TempDir); err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	if err := os.MkdirAll(config.Storage.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize components
	log.Println("Initializing components...")

	// Whisper transcriber
	transcriber, err := transcription.NewWhisperTranscriber(
		config.Whisper.ModelPath,
		config.Whisper.Threads,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Whisper: %v", err)
	}

	// Local storage
	localStorage := storage.NewLocalStorage(config.Storage.OutputDir)

	// Google Drive client (optional - may fail if credentials not set up)
	var driveClient *storage.DriveClient
	if _, err := os.Stat(config.GoogleDrive.CredentialsFile); err == nil {
		driveClient, err = storage.NewDriveClient(
			config.GoogleDrive.CredentialsFile,
			config.GoogleDrive.TokenFile,
			config.GoogleDrive.FolderName,
		)
		if err != nil {
			log.Printf("WARNING: Google Drive not available: %v", err)
			log.Println("Transcripts will only be saved locally")
			driveClient = nil
		} else {
			log.Println("Google Drive integration enabled")
		}
	} else {
		log.Println("Google Drive credentials not found - saving locally only")
	}

	// Database
	db, err := storage.NewMetadataDB(config.Storage.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Worker pool
	workerPool := queue.NewWorkerPool(
		config.Workers.Count,
		transcriber,
		localStorage,
		driveClient,
		db,
	)
	workerPool.Start()

	// Cleanup scheduler
	cleanupScheduler := cleanup.NewScheduler(
		config.Storage.TempDir,
		config.Cleanup.IntervalMinutes,
		config.Cleanup.MaxAgeHours,
	)
	cleanupScheduler.Start()
	defer cleanupScheduler.Stop()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: config.Limits.MaxFileSizeMB * 1024 * 1024,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(workerPool, config.Limits.MaxFileSizeMB)
	gdriveHandler := handlers.NewGDriveHandler(workerPool)
	youtubeHandler := handlers.NewYouTubeHandler(workerPool)
	streamHandler := handlers.NewStreamHandler(workerPool)

	// Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"version": "1.0.0",
		})
	})

	app.Post("/upload", uploadHandler.Handle)
	app.Post("/gdrive", gdriveHandler.Handle)
	app.Post("/youtube", youtubeHandler.Handle)

	// WebSocket route
	app.Get("/ws/stream", websocket.New(streamHandler.Handle))

	// Get transcript metadata
	app.Get("/transcripts", func(c *fiber.Ctx) error {
		limit := 50 // Default limit
		transcripts, err := db.ListTranscripts(limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(transcripts)
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	log.Println("📝 Endpoints:")
	log.Println("   POST /upload      - Upload audio file")
	log.Println("   POST /gdrive      - Process Google Drive link")
	log.Println("   POST /youtube     - Capture YouTube audio")
	log.Println("   GET  /ws/stream   - WebSocket audio streaming")
	log.Println("   GET  /transcripts - List all transcripts")
	log.Println("   GET  /health      - Health check")

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down gracefully...")
		app.Shutdown()
	}()

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// loadConfig loads configuration from YAML file
func loadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
