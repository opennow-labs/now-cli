//go:build linux

package detect

import (
	"os/exec"
	"strings"
)

// detectMedia checks for playing media via playerctl (MPRIS D-Bus) on Linux.
func detectMedia() MediaResult {
	// Check if anything is playing
	out, err := exec.Command("playerctl", "status").Output()
	if err != nil || strings.TrimSpace(string(out)) != "Playing" {
		return MediaResult{}
	}

	// Get track title
	out, err = exec.Command("playerctl", "metadata", "title").Output()
	if err != nil {
		return MediaResult{}
	}
	title := strings.TrimSpace(string(out))
	if title == "" {
		return MediaResult{}
	}

	var artist, album, playerName string

	out, err = exec.Command("playerctl", "metadata", "artist").Output()
	if err == nil {
		artist = strings.TrimSpace(string(out))
	}

	out, err = exec.Command("playerctl", "metadata", "xesam:album").Output()
	if err == nil {
		album = strings.TrimSpace(string(out))
	}

	// Get player name (e.g., "spotify", "vlc", "chromium")
	out, err = exec.Command("playerctl", "-l").Output()
	if err == nil {
		// playerctl -l returns one player per line; first is the active one
		players := strings.TrimSpace(string(out))
		if first, _, ok := strings.Cut(players, "\n"); ok {
			playerName = first
		} else {
			playerName = players
		}
		// Strip instance suffix (e.g., "chromium.instance1234" → "chromium")
		if dot := strings.Index(playerName, "."); dot > 0 {
			playerName = playerName[:dot]
		}
	}

	return ClassifyMedia(&MediaInfo{
		Title:     title,
		Artist:    artist,
		Album:     album,
		SourceID:  playerName,
		IsPlaying: true,
	})
}
