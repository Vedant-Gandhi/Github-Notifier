package main

import (
	"context"
	"flag"
	"gitnotifier/config"
	"gitnotifier/internal/github"
	"gitnotifier/internal/notifier"
	"gitnotifier/internal/repository"
	"gitnotifier/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Add command line flag for env file path
	envFile := flag.String("env", "", "Path to environment file")
	flag.Parse()

	// Load environment file if specified, otherwise try default .env
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			log.Fatalf("Error loading environment file %s: %v", *envFile, err)
		}
	} else if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Rest of the code remains the same
	repoURL := os.Getenv("GITHUB_REPO_URL")
	if repoURL == "" {
		log.Fatal("GITHUB_REPO_URL environment variable is not set")
	}

	// Parse GitHub repository URL
	owner, repo, err := github.ParseGitHubURL(repoURL)
	if err != nil {
		log.Fatalf("Invalid repository URL: %v", err)
	}

	// Get poll interval from environment
	pollInterval := config.DefaultPollInterval
	if envInterval := os.Getenv("POLL_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			if d < config.MinPollInterval {
				d = config.MinPollInterval
			}
			pollInterval = d
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: config.HTTPTimeout,
	}

	// Initialize repository
	githubRepo := repository.NewRepository(
		client,
		owner,
		repo,
		os.Getenv("GITHUB_TOKEN"),
	)

	// Initialize platform-specific notifier
	notifier, err := notifier.NewPlatformNotifier()
	if err != nil {
		log.Fatalf("Failed to initialize notifier: %v", err)
	}

	// Create notification service
	service := service.NewService(githubRepo, notifier, pollInterval)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating shutdown...", sig)
		cancel()
	}()

	// Start the service
	if err := service.Start(ctx); err != nil {
		log.Fatalf("Service error: %v", err)
	}
}
