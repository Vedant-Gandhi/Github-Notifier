package platform

import (
	"fmt"
	"os/exec"

	"github.com/gen2brain/beeep"
)

// LinuxNotifier implements desktop notifications for Linux
type LinuxNotifier struct {
}

func NewLinuxNotifier() *LinuxNotifier {
	return &LinuxNotifier{}
}

func (n *LinuxNotifier) Notify(title, message, url string) error {
	cmd := exec.Command("notify-send", title, fmt.Sprintf("%s\n%s", message, url))
	if err := cmd.Run(); err != nil {
		// Fall back to beeep if native notifications fail
		return beeep.Notify(title, fmt.Sprintf("%s\n\nClick to open: %s", message, url), "")
	}
	return nil
}
