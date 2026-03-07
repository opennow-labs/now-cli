//go:build linux

package detect

import (
	"os/exec"
	"strings"
)

// detectMusic checks for playing music via playerctl (MPRIS D-Bus) on Linux.
func detectMusic() (artist, track string) {
	// Check if anything is playing
	out, err := exec.Command("playerctl", "status").Output()
	if err != nil || strings.TrimSpace(string(out)) != "Playing" {
		return "", ""
	}

	// Get track title
	out, err = exec.Command("playerctl", "metadata", "title").Output()
	if err != nil {
		return "", ""
	}
	track = strings.TrimSpace(string(out))

	// Get artist
	out, err = exec.Command("playerctl", "metadata", "artist").Output()
	if err == nil {
		artist = strings.TrimSpace(string(out))
	}

	return artist, track
}
