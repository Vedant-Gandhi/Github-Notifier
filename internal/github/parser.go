package github

import (
	"fmt"
	"strings"
)

// ParseGitHubURL parses a GitHub repository URL into owner and repo parts
// Only accepts full GitHub URLs in the format: https://github.com/owner/repo
func ParseGitHubURL(url string) (owner, repo string, err error) {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/issues")

	if !strings.HasPrefix(url, "https://github.com/") {
		return "", "", fmt.Errorf("invalid GitHub URL format. URL must start with 'https://github.com/'")
	}

	// Remove the prefix to get owner/repo part
	repoPath := strings.TrimPrefix(url, "https://github.com/")
	parts := strings.Split(repoPath, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format. Expected 'https://github.com/owner/repo'")
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("owner and repo cannot be empty")
	}

	return parts[0], parts[1], nil
}
