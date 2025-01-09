package platform

import (
	"fmt"
	"os/exec"
)

// MacOSNotifier implements desktop notifications for macOS
type MacOSNotifier struct{}

func NewMacOSNotifier() *MacOSNotifier {
	return &MacOSNotifier{}
}

func (n *MacOSNotifier) Notify(title, message, url string) error {
	cmd := exec.Command("terminal-notifier",
		"-title", title,
		"-message", message,
		"-open", url,
		"-sound", "default")

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("terminal-notifier not installed. Please install with: brew install terminal-notifier")
		}
		return err
	}
	return nil
}
