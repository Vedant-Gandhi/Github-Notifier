package notifier

import (
	"fmt"
	"gitnotifier/internal/issue"
	"gitnotifier/internal/notifier/platform"
	"runtime"
)

// NotificationMessage represents a notification to be sent
type NotificationMessage struct {
	Title   string
	Message string
	URL     string
}

// IssueNotifier converts issues to notification messages
type IssueNotifier struct {
	notifier Notifier
}

// NewIssueNotifier creates a new IssueNotifier
func NewIssueNotifier(notifier Notifier) *IssueNotifier {
	return &IssueNotifier{
		notifier: notifier,
	}
}

// NotifyNewIssue sends a notification for a new issue
func (in *IssueNotifier) NotifyNewIssue(issue issue.Issue) error {
	title := "New GitHub Issue"
	message := formatIssueMessage(issue)
	return in.notifier.Notify(title, message, issue.HTMLURL)
}

func formatIssueMessage(issue issue.Issue) string {
	return fmt.Sprintf("#%d: %s", issue.Number, issue.Title)
}

// NewPlatformNotifier creates the appropriate notifier for the current platform
func NewPlatformNotifier() (Notifier, error) {
	switch runtime.GOOS {
	case "darwin":
		return platform.NewMacOSNotifier(), nil
	case "windows":
		return platform.NewWindowsNotifier(), nil
	case "linux":
		return platform.NewLinuxNotifier(), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
