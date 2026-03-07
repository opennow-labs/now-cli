package template

import (
	"testing"

	"github.com/nownow-labs/nownow/internal/detect"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		ctx      detect.Context
		activity string
		want     string
	}{
		{
			name:     "app with activity",
			tmpl:     "{activity}",
			ctx:      detect.Context{App: "VS Code"},
			activity: "Coding",
			want:     "Coding",
		},
		{
			name:     "activity with app",
			tmpl:     "{activity} · {app}",
			ctx:      detect.Context{App: "VS Code"},
			activity: "Coding",
			want:     "Coding · VS Code",
		},
		{
			name:     "music template",
			tmpl:     "{app} · {music}",
			ctx:      detect.Context{App: "Spotify", MusicArtist: "Daft Punk", MusicTrack: "Get Lucky"},
			activity: "",
			want:     "Spotify · Daft Punk - Get Lucky",
		},
		{
			name:     "empty context",
			tmpl:     "{activity}",
			ctx:      detect.Context{},
			activity: "",
			want:     "",
		},
		{
			name:     "activity only",
			tmpl:     "{activity}",
			ctx:      detect.Context{App: "Terminal"},
			activity: "In terminal",
			want:     "In terminal",
		},
		{
			name:     "custom template with title",
			tmpl:     "{app}: {title}",
			ctx:      detect.Context{App: "Chrome", WindowTitle: "GitHub - nownow-labs/nownow"},
			activity: "",
			want:     "Chrome: GitHub - nownow-labs/nownow",
		},
		{
			name:     "watching template",
			tmpl:     "{activity}",
			ctx:      detect.Context{Watching: "Stranger Things"},
			activity: "Watching",
			want:     "Watching",
		},
		{
			name:     "watching empty",
			tmpl:     "{app} · {watching}",
			ctx:      detect.Context{App: "Chrome"},
			activity: "",
			want:     "Chrome",
		},
		{
			name:     "music subfields",
			tmpl:     "{music.artist} playing {music.track}",
			ctx:      detect.Context{MusicArtist: "Queen", MusicTrack: "Radio Ga Ga"},
			activity: "",
			want:     "Queen playing Radio Ga Ga",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.tmpl, tt.ctx, tt.activity)
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
