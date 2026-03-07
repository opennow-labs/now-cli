//go:build linux

package detect

import (
	"os/exec"
	"strings"
)

// detectApp returns the focused application name and window title on Linux via xdotool.
func detectApp() (app, title string) {
	// Get active window ID
	out, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return "", ""
	}
	winID := strings.TrimSpace(string(out))
	if winID == "" {
		return "", ""
	}

	// Get window name (title)
	out, err = exec.Command("xdotool", "getactivewindow", "getwindowname").Output()
	if err == nil {
		title = strings.TrimSpace(string(out))
	}

	// Get WM_CLASS for the app name
	out, err = exec.Command("xprop", "-id", winID, "WM_CLASS").Output()
	if err == nil {
		// WM_CLASS returns: WM_CLASS(STRING) = "instance", "Class"
		s := string(out)
		if idx := strings.LastIndex(s, `"`); idx > 0 {
			s = s[:idx]
			if idx2 := strings.LastIndex(s, `"`); idx2 >= 0 {
				app = s[idx2+1:]
			}
		}
	}

	return app, title
}
