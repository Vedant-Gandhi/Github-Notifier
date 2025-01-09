package config

import "time"

// Constants for service configuration
const (
	MaxNotificationLength = 100
	MinPollInterval       = 1 * time.Minute
	DefaultPollInterval   = 5 * time.Minute
	MaxRetries            = 3
	RetryDelay            = 5 * time.Second
	HTTPTimeout           = 10 * time.Second
	NotifyDelay           = 500 * time.Millisecond // Prevent notification flooding
)
