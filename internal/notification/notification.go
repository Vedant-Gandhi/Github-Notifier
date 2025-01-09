package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"gitnotifier/config"
	"gitnotifier/internal/issue"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gen2brain/beeep"
	"golang.org/x/time/rate"
)

// Service handles GitHub issue notifications
type Service struct {
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

// NewService creates a new notification service
func NewService(owner, repo string, pollInterval time.Duration) *Service {
	return &Service{
		lastCheckID:  0,
		pollInterval: pollInterval,
		limiter:      rate.NewLimiter(rate.Every(time.Minute), 30),
		client:       &http.Client{Timeout: config.HTTPTimeout},
		shutdownChan: make(chan struct{}),
		maxRetries:   config.MaxRetries,
		owner:        owner,
		repo:         repo,
	}
}

func (s *Service) fetchLatestIssues(ctx context.Context) ([]issue.Issue, error) {
	if err := s.limiter.Wait(ctx); err != nil {
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

	var issues []issue.Issue
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		resp, err := s.client.Do(req)
		if err != nil {
			if attempt == s.maxRetries {
				return nil, fmt.Errorf("error fetching issues after %d attempts: %v", attempt, err)
			}
			time.Sleep(config.RetryDelay)
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
			time.Sleep(config.RetryDelay)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}
		break
	}

	return issues, nil
}

func (s *Service) sendNotification(issue issue.Issue) error {
	s.notifyMutex.Lock()
	defer s.notifyMutex.Unlock()

	if time.Since(s.lastNotifyTime) < config.NotifyDelay {
		time.Sleep(config.NotifyDelay)
	}

	title := "New GitHub Issue"
	message := fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
	if len(message) > config.MaxNotificationLength {
		message = message[:config.MaxNotificationLength-3] + "..."
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
		cmd := exec.Command("notify-send", title, message)
		if err = cmd.Run(); err != nil {
			err = beeep.Notify(title, message+"\n\nClick to open: "+issue.HTMLURL, "")
		}
	default:
		err = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	s.lastNotifyTime = time.Now()
	return err
}

func (s *Service) checkForNewIssues(ctx context.Context) error {
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

// Start begins the notification service
func (s *Service) Start(ctx context.Context) error {
	log.Printf("Starting GitHub issues notification service for %s/%s...", s.owner, s.repo)
	log.Printf("Poll interval: %v", s.pollInterval)

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
		case <-ctx.Done():
			log.Println("Context cancelled, stopping service...")
			return nil
		case <-s.shutdownChan:
			log.Println("Shutdown requested, stopping service...")
			return nil
		}
	}
}

// Stop gracefully stops the notification service
func (s *Service) Stop() {
	close(s.shutdownChan)
}
