//go:build darwin

package detect

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type nowPlayingJSON struct {
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	SourceID  string `json:"source_id"`
	IsPlaying bool   `json:"is_playing"`
}

// helperName is the Swift CLI helper binary name.
const helperName = "nowplaying-helper"

// findHelper locates the nowplaying-helper binary next to the current executable.
func findHelper() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	p := filepath.Join(filepath.Dir(exe), helperName)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// detectMediaViaHelper calls the Swift helper and returns MediaInfo.
func detectMediaViaHelper() *MediaInfo {
	helper := findHelper()
	if helper == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, helper).Output()
	if err != nil {
		return nil
	}

	var result nowPlayingJSON
	if err := json.Unmarshal(out, &result); err != nil {
		return nil
	}

	return &MediaInfo{
		Title:     result.Title,
		Artist:    result.Artist,
		Album:     result.Album,
		SourceID:  result.SourceID,
		IsPlaying: result.IsPlaying,
	}
}

// detectMedia tries the Swift helper first, then falls back to AppleScript.
func detectMedia() MediaResult {
	if info := detectMediaViaHelper(); info != nil {
		return ClassifyMedia(info)
	}

	// Fallback: existing AppleScript-based detection (music only)
	artist, track := detectSpotify()
	if track != "" {
		return MediaResult{MusicArtist: artist, MusicTrack: track}
	}
	artist, track = detectAppleMusic()
	if track != "" {
		return MediaResult{MusicArtist: artist, MusicTrack: track}
	}
	return MediaResult{}
}
