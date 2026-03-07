package template

import (
	"testing"

	"github.com/ctx-st/nownow/internal/detect"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name  string
		tmpl  string
		ctx   detect.Context
		emoji string
		want  string
	}{
		{
			name:  "full context",
			tmpl:  "{emoji} {app} · {project} ({branch})",
			ctx:   detect.Context{App: "VS Code", Project: "nownow", Branch: "main"},
			emoji: "\U0001F4BB",
			want:  "\U0001F4BB VS Code · nownow (main)",
		},
		{
			name:  "no branch",
			tmpl:  "{emoji} {app} · {project} ({branch})",
			ctx:   detect.Context{App: "Figma", Project: "design"},
			emoji: "\U0001F3A8",
			want:  "\U0001F3A8 Figma · design",
		},
		{
			name:  "no project no branch",
			tmpl:  "{emoji} {app} · {project} ({branch})",
			ctx:   detect.Context{App: "Safari"},
			emoji: "\U0001F310",
			want:  "\U0001F310 Safari",
		},
		{
			name:  "music template",
			tmpl:  "{app} · {music}",
			ctx:   detect.Context{App: "Spotify", MusicArtist: "Daft Punk", MusicTrack: "Get Lucky"},
			emoji: "",
			want:  "Spotify · Daft Punk - Get Lucky",
		},
		{
			name:  "empty context",
			tmpl:  "{emoji} {app} · {project} ({branch})",
			ctx:   detect.Context{},
			emoji: "",
			want:  "",
		},
		{
			name:  "only emoji",
			tmpl:  "{emoji} {app}",
			ctx:   detect.Context{App: "Terminal"},
			emoji: "\u26A1",
			want:  "\u26A1 Terminal",
		},
		{
			name:  "custom template with title",
			tmpl:  "{app}: {title}",
			ctx:   detect.Context{App: "Chrome", WindowTitle: "GitHub - ctx-st/nownow"},
			emoji: "",
			want:  "Chrome: GitHub - ctx-st/nownow",
		},
		{
			name:  "music subfields",
			tmpl:  "{music.artist} playing {music.track}",
			ctx:   detect.Context{MusicArtist: "Queen", MusicTrack: "Radio Ga Ga"},
			emoji: "",
			want:  "Queen playing Radio Ga Ga",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.tmpl, tt.ctx, tt.emoji)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCleanSeparators(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello () world", "hello world"},
		{"a · · b", "a · b"},
		{"  extra  spaces  ", "extra spaces"},
		{"trailing ·", "trailing"},
		{"· leading", "leading"},
	}

	for _, tt := range tests {
		got := cleanSeparators(tt.input)
		if got != tt.want {
			t.Errorf("cleanSeparators(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
