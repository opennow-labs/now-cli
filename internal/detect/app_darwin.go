//go:build darwin

package detect

import (
	"os/exec"
	"strings"
)

// detectApp returns the frontmost application name and its window title on macOS.
// Uses lsappinfo which works reliably from background processes (unlike osascript).
func detectApp() (app, title string) {
	// Get frontmost app ASN
	front, err := exec.Command("lsappinfo", "front").Output()
	if err != nil {
		return "", ""
	}
	asn := strings.TrimSpace(string(front))
	if asn == "" {
		return "", ""
	}

	// Get app name from ASN
	out, err := exec.Command("lsappinfo", "info", "-only", "name", asn).Output()
	if err != nil {
		return "", ""
	}
	// Output: "LSDisplayName"="App Name"
	app = parseLsappinfoValue(string(out))

	// Window title via osascript (best effort — may fail from background)
	out, err = exec.Command("osascript", "-e",
		`tell application "System Events" to get title of front window of (first application process whose frontmost is true)`).Output()
	if err == nil {
		title = strings.TrimSpace(string(out))
	}

	return app, title
}

// parseLsappinfoValue extracts the value from "Key"="Value" format.
func parseLsappinfoValue(s string) string {
	s = strings.TrimSpace(s)
	idx := strings.Index(s, "=\"")
	if idx < 0 {
		return ""
	}
	val := s[idx+2:]
	if strings.HasSuffix(val, "\"") {
		val = val[:len(val)-1]
	}
	return val
}
