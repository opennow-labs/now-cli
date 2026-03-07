package detect

import (
	"testing"
)

func TestContextMusic(t *testing.T) {
	tests := []struct {
		name     string
		ctx      Context
		wantStr  string
		wantBool bool
	}{
		{
			name:     "no music",
			ctx:      Context{},
			wantStr:  "",
			wantBool: false,
		},
		{
			name:     "track only",
			ctx:      Context{MusicTrack: "Bohemian Rhapsody"},
			wantStr:  "Bohemian Rhapsody",
			wantBool: true,
		},
		{
			name:     "artist and track",
			ctx:      Context{MusicArtist: "Queen", MusicTrack: "Bohemian Rhapsody"},
			wantStr:  "Queen - Bohemian Rhapsody",
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ctx.HasMusic(); got != tt.wantBool {
				t.Errorf("HasMusic() = %v, want %v", got, tt.wantBool)
			}
			if got := tt.ctx.Music(); got != tt.wantStr {
				t.Errorf("Music() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestDetectReturnsContext(t *testing.T) {
	// Detect should never panic, even if nothing is available
	ctx := Detect()
	// Just verify it returns a valid struct (fields may be empty)
	_ = ctx.App
	_ = ctx.Project
	_ = ctx.Branch
	_ = ctx.Watching
}

func TestDetectWatching(t *testing.T) {
	tests := []struct {
		name        string
		windowTitle string
		want        string
	}{
		{"YouTube video", "Never Gonna Give You Up - YouTube", "Never Gonna Give You Up"},
		{"Netflix show", "The Office S3E5 - Netflix", "The Office S3E5"},
		{"Twitch stream", "shroud - Twitch", "shroud"},
		{"Bilibili with dash", "Some Video - Bilibili", "Some Video"},
		{"Bilibili with underscore", "Some Video _ Bilibili", "Some Video"},
		{"Disney+", "The Mandalorian - Disney+", "The Mandalorian"},
		{"Prime Video", "The Boys - Prime Video", "The Boys"},
		{"no match", "main.go - nownow", ""},
		{"empty title", "", ""},
		{"suffix only", " - YouTube", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectWatching(tt.windowTitle)
			if got != tt.want {
				t.Errorf("detectWatching(%q) = %q, want %q", tt.windowTitle, got, tt.want)
			}
		})
	}
}

func TestHasWatching(t *testing.T) {
	ctx := Context{}
	if ctx.HasWatching() {
		t.Error("HasWatching() should be false for empty context")
	}
	ctx.Watching = "The Office"
	if !ctx.HasWatching() {
		t.Error("HasWatching() should be true when Watching is set")
	}
}
