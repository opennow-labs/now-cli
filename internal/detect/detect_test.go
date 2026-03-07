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
	_ = ctx.Watching
}

func TestDetectWatching(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"YouTube", "Rick Astley - Never Gonna Give You Up - YouTube", "Rick Astley - Never Gonna Give You Up"},
		{"Netflix", "Stranger Things - Netflix", "Stranger Things"},
		{"Twitch", "shroud - Live - Twitch", "shroud - Live"},
		{"Disney+", "The Mandalorian - Disney+", "The Mandalorian"},
		{"Prime Video", "The Boys - Prime Video", "The Boys"},
		{"Bilibili underscore", "some video _ Bilibili", "some video"},
		{"Bilibili dash", "some video - Bilibili", "some video"},
		{"no match", "Visual Studio Code", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectWatching(tt.title)
			if got != tt.want {
				t.Errorf("detectWatching(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestHasWatching(t *testing.T) {
	ctx := Context{}
	if ctx.HasWatching() {
		t.Error("expected HasWatching() = false for empty context")
	}
	ctx.Watching = "Some Video"
	if !ctx.HasWatching() {
		t.Error("expected HasWatching() = true when Watching is set")
	}
}
