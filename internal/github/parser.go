package github

import (
	"fmt"
	"strings"
)

// ParseGitHubURL parses a GitHub repository URL into owner and repo parts
func ParseGitHubURL(url string) (owner, repo string, err error) {
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
