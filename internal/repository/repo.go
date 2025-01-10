package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"gitnotifier/internal/issue"
	"net/http"
)

// IssueRepository defines the interface for fetching issues
type IssueRepository interface {
	FetchLatestIssues(ctx context.Context) ([]issue.Issue, error)
}

// Repository implements GitHub API communication
type Repository struct {
	client *http.Client
	owner  string
	repo   string
	token  string
}

// NewRepository creates a new GitHub repository client
func NewRepository(client *http.Client, owner, repo, token string) *Repository {
	return &Repository{
		client: client,
		owner:  owner,
		repo:   repo,
		token:  token,
	}
}

// FetchLatestIssues fetches the latest issues (excluding pull requests) from GitHub
func (r *Repository) FetchLatestIssues(ctx context.Context) ([]issue.Issue, error) {
	// Note the addition of `is:issue` to exclude pull requests
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open&sort=created&direction=desc&per_page=10&is=issue",
		r.owner, r.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if r.token != "" {
		req.Header.Add("Authorization", "Bearer "+r.token)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("User-Agent", "GitHub-Issue-Notifier")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching issues: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("GitHub API authentication failed. Please check your token")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status code: %d", resp.StatusCode)
	}

	var issues []issue.Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Additional check to filter out any pull requests that might have slipped through
	var filteredIssues []issue.Issue
	for _, issue := range issues {
		// GitHub Pull Requests have a "pull_request" field
		if issue.PullRequest == nil {
			filteredIssues = append(filteredIssues, issue)
		}
	}

	return filteredIssues, nil
}
