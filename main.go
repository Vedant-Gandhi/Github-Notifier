package main

import (
	"context"
	"gitnotifier/config"
	"gitnotifier/internal/github"
	"gitnotifier/internal/notification"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Get repository URL from environment
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

	// Create notification service
	service := notification.NewService(owner, repo, pollInterval)

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
