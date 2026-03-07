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
}
