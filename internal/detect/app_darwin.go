//go:build darwin

package detect

import (
	"os/exec"
	"strings"
)

// detectApp returns the frontmost application name and its window title on macOS.
func detectApp() (app, title string) {
	// Get frontmost app name
	out, err := exec.Command("osascript", "-e",
		`tell application "System Events" to get name of first application process whose frontmost is true`).Output()
	if err != nil {
		return "", ""
	}
	app = strings.TrimSpace(string(out))

	// Get window title
	out, err = exec.Command("osascript", "-e",
		`tell application "System Events" to get title of front window of (first application process whose frontmost is true)`).Output()
	if err != nil {
		// App may not have a window — that's fine
		return app, ""
	}
	title = strings.TrimSpace(string(out))

	return app, title
}
