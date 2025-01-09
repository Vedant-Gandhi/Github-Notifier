package notifier

// Notifier defines the interface for sending notifications
type Notifier interface {
	Notify(title string, message string, url string) error
}
