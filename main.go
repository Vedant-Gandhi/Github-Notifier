package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type Issue struct {
	ID        int       `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	HTMLURL   string    `json:"html_url"`
	State     string    `json:"state"`
}

type NotificationService struct {
	lastCheckID    int
	pollInterval   time.Duration
	limiter        *rate.Limiter
	client         *http.Client
	shutdownChan   chan struct{}
	wg             sync.WaitGroup
	lastNotifyTime time.Time
	notifyMutex    sync.Mutex
	retryAttempts  int
	maxRetries     int
	owner          string
	repo           string
}

const (
	maxNotificationLength = 100
	minPollInterval       = 1 * time.Minute
	defaultPollInterval   = 5 * time.Minute
	maxRetries            = 3
	retryDelay            = 5 * time.Second
	httpTimeout           = 10 * time.Second
	notifyDelay           = 500 * time.Millisecond // Prevent notification flooding
)

func NewNotificationService() (*NotificationService, error) {
	// Parse repository URL
	repoURL := os.Getenv("GITHUB_REPO_URL")
	if repoURL == "" {
		return nil, fmt.Errorf("GITHUB_REPO_URL environment variable is not set")
	}

	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %v", err)
	}

	// Validate and adjust poll interval
	pollInterval := defaultPollInterval
	if envInterval := os.Getenv("POLL_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			if d < minPollInterval {
				d = minPollInterval
			}
			pollInterval = d
		}
	}

	return &NotificationService{
		lastCheckID:  0,
		pollInterval: pollInterval,
		limiter:      rate.NewLimiter(rate.Every(time.Minute), 30), // GitHub API rate limit
		client:       &http.Client{Timeout: httpTimeout},
		shutdownChan: make(chan struct{}),
		maxRetries:   maxRetries,
		owner:        owner,
		repo:         repo,
	}, nil
}

func parseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/issues")

	if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimPrefix(url, "https://github.com/")
	}

	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format. Expected 'owner/repo' or 'https://github.com/owner/repo'")
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("owner and repo cannot be empty")
	}

	return parts[0], parts[1], nil
}

func (s *NotificationService) fetchLatestIssues(ctx context.Context) ([]Issue, error) {
	// Respect rate limiting
	err := s.limiter.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("rate limit error: %v", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open&sort=created&direction=desc&per_page=10",
		s.owner, s.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("User-Agent", "GitHub-Issue-Notifier")

	var issues []Issue
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		resp, err := s.client.Do(req)
		if err != nil {
			if attempt == s.maxRetries {
				return nil, fmt.Errorf("error fetching issues after %d attempts: %v", attempt, err)
			}
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("GitHub API authentication failed. Please check your token")
		}

		if resp.StatusCode != http.StatusOK {
			if attempt == s.maxRetries {
				return nil, fmt.Errorf("GitHub API returned status code: %d", resp.StatusCode)
			}
			time.Sleep(retryDelay)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}
		break
	}

	return issues, nil
}

func (s *NotificationService) sendNotification(issue Issue) error {
	s.notifyMutex.Lock()
	defer s.notifyMutex.Unlock()

	// Prevent notification flooding
	if time.Since(s.lastNotifyTime) < notifyDelay {
		time.Sleep(notifyDelay)
	}

	title := "New GitHub Issue"
	message := fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
	if len(message) > maxNotificationLength {
		message = message[:maxNotificationLength-3] + "..."
	}

	var err error
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("terminal-notifier",
			"-title", title,
			"-message", message,
			"-open", issue.HTMLURL,
			"-sound", "default")

		if err = cmd.Run(); err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("terminal-notifier not installed. Please install with: brew install terminal-notifier")
			}
		}
	case "windows":
		err = beeep.Notify(title, message+"\n\nClick to open: "+issue.HTMLURL, "")
	case "linux":
		// Try native notifications first
		cmd := exec.Command("notify-send", title, message)
		if err = cmd.Run(); err != nil {
			// Fall back to beeep if native notifications fail
			err = beeep.Notify(title, message+"\n\nClick to open: "+issue.HTMLURL, "")
		}
	default:
		err = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	s.lastNotifyTime = time.Now()
	return err
}

func (s *NotificationService) checkForNewIssues(ctx context.Context) error {
	issues, err := s.fetchLatestIssues(ctx)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		if issue.ID > s.lastCheckID {
			if err := s.sendNotification(issue); err != nil {
				log.Printf("Error sending notification for issue #%d: %v", issue.Number, err)
				continue
			}
			log.Printf("Sent notification for new issue #%d: %s", issue.Number, issue.Title)

			if issue.ID > s.lastCheckID {
				s.lastCheckID = issue.ID
			}
		}
	}

	return nil
}

func (s *NotificationService) Start() error {
	log.Printf("Starting GitHub issues notification service for %s/%s...", s.owner, s.repo)
	log.Printf("Poll interval: %v", s.pollInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initial check
	if err := s.checkForNewIssues(ctx); err != nil {
		log.Printf("Error during initial check: %v", err)
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.checkForNewIssues(ctx); err != nil {
				log.Printf("Error checking for new issues: %v", err)
			}
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			return nil
		case <-s.shutdownChan:
			log.Println("Shutdown requested, stopping service...")
			return nil
		}
	}
}

func (s *NotificationService) Stop() {
	close(s.shutdownChan)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	service, err := NewNotificationService()
	if err != nil {
		log.Fatalf("Error initializing service: %v", err)
	}

	if err := service.Start(); err != nil {
		log.Fatalf("Service error: %v", err)
	}
}
