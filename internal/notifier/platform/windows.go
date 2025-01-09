package platform

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

// WindowsNotifier implements desktop notifications for Windows
type WindowsNotifier struct{}

func NewWindowsNotifier() *WindowsNotifier {
	return &WindowsNotifier{}
}

func (n *WindowsNotifier) Notify(title, message, url string) error {
	return beeep.Notify(title, fmt.Sprintf("%s\n\nClick to open: %s", message, url), "")
}
