package service

import (
	"context"
	"fmt"
	"gitnotifier/internal/notifier"
	"gitnotifier/internal/repository"
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Service handles GitHub issue monitoring and notifications
type Service struct {
	repo           repository.IssueRepository
	issueNotifier  *notifier.IssueNotifier
	lastCheckID    int
	pollInterval   time.Duration
	limiter        *rate.Limiter
	shutdownChan   chan struct{}
	wg             sync.WaitGroup
	lastNotifyTime time.Time
	notifyMutex    sync.Mutex
}

// NewService creates a new notification service
func NewService(repo repository.IssueRepository, n notifier.Notifier, pollInterval time.Duration) *Service {
	return &Service{
		repo:          repo,
		issueNotifier: notifier.NewIssueNotifier(n),
		pollInterval:  pollInterval,
		limiter:       rate.NewLimiter(rate.Every(time.Minute), 30),
		shutdownChan:  make(chan struct{}),
	}
}

func (s *Service) checkForNewIssues(ctx context.Context) error {
	// Respect rate limiting
	if err := s.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %v", err)
	}

	issues, err := s.repo.FetchLatestIssues(ctx)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		if issue.ID > s.lastCheckID {
			if err := s.issueNotifier.NotifyNewIssue(issue); err != nil {
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
	log.Printf("Starting GitHub issues notification service...")
	log.Printf("Poll interval: %v", s.pollInterval)

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
