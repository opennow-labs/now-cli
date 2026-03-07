//go:build darwin

package detect

import (
	"os/exec"
	"strings"
)

func detectSpotify() (artist, track string) {
	// Check if Spotify is running
	out, err := exec.Command("osascript", "-e",
		`tell application "System Events" to (name of processes) contains "Spotify"`).Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return "", ""
	}

	// Check player state
	out, err = exec.Command("osascript", "-e",
		`tell application "Spotify" to player state as string`).Output()
	if err != nil || strings.TrimSpace(string(out)) != "playing" {
		return "", ""
	}

	// Get track info
	out, err = exec.Command("osascript", "-e",
		`tell application "Spotify" to name of current track`).Output()
	if err != nil {
		return "", ""
	}
	track = strings.TrimSpace(string(out))

	out, err = exec.Command("osascript", "-e",
		`tell application "Spotify" to artist of current track`).Output()
	if err != nil {
		return "", track
	}
	artist = strings.TrimSpace(string(out))

	return artist, track
}

func detectAppleMusic() (artist, track string) {
	// Check if Music is running
	out, err := exec.Command("osascript", "-e",
		`tell application "System Events" to (name of processes) contains "Music"`).Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return "", ""
	}

	// Check player state
	out, err = exec.Command("osascript", "-e",
		`tell application "Music" to player state as string`).Output()
	if err != nil || strings.TrimSpace(string(out)) != "playing" {
		return "", ""
	}

	// Get track info
	out, err = exec.Command("osascript", "-e",
		`tell application "Music" to name of current track`).Output()
	if err != nil {
		return "", ""
	}
	track = strings.TrimSpace(string(out))

	out, err = exec.Command("osascript", "-e",
		`tell application "Music" to artist of current track`).Output()
	if err != nil {
		return "", track
	}
	artist = strings.TrimSpace(string(out))

	return artist, track
}
