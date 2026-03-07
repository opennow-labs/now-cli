package template

import (
	"testing"

	"github.com/nownow-labs/nownow/internal/detect"
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
			name:  "app with emoji",
			tmpl:  "{emoji} {app}",
			ctx:   detect.Context{App: "VS Code"},
			emoji: "\U0001F4BB",
			want:  "\U0001F4BB VS Code",
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
			tmpl:  "{emoji} {app}",
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
			ctx:   detect.Context{App: "Chrome", WindowTitle: "GitHub - nownow-labs/nownow"},
			emoji: "",
			want:  "Chrome: GitHub - nownow-labs/nownow",
		},
		{
			name:  "watching template",
			tmpl:  "{emoji} watching: {watching}",
			ctx:   detect.Context{Watching: "Stranger Things"},
			emoji: "\U0001F4FA",
			want:  "\U0001F4FA watching: Stranger Things",
		},
		{
			name:  "watching empty",
			tmpl:  "{app} · {watching}",
			ctx:   detect.Context{App: "Chrome"},
			emoji: "",
			want:  "Chrome",
		},
		{
			name:  "music subfields",
			tmpl:  "{music.artist} playing {music.track}",
			ctx:   detect.Context{MusicArtist: "Queen", MusicTrack: "Radio Ga Ga"},
			emoji: "",
			want:  "Queen playing Radio Ga Ga",
		},
		{
			name:  "legacy project/branch placeholders render as literals",
			tmpl:  "{emoji} {app} · {project} ({branch})",
			ctx:   detect.Context{App: "VS Code"},
			emoji: "\U0001F4BB",
			want:  "\U0001F4BB VS Code · {project} ({branch})",
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
