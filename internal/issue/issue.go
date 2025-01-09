package issue

import "time"

// Issue represents a GitHub issue
type Issue struct {
	ID        int       `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	HTMLURL   string    `json:"html_url"`
	State     string    `json:"state"`
}
