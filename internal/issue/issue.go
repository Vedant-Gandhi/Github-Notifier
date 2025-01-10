package issue

import "time"

// PullRequest represents the pull_request field in GitHub's API
type PullRequest struct {
	URL string `json:"url"`
}

// Issue represents a GitHub issue
type Issue struct {
	ID          int          `json:"id"`
	Number      int          `json:"number"`
	Title       string       `json:"title"`
	CreatedAt   time.Time    `json:"created_at"`
	HTMLURL     string       `json:"html_url"`
	State       string       `json:"state"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`
}
