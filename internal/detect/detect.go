package detect

import "strings"

// Context holds the detected environment context.
type Context struct {
	App         string `json:"app,omitempty"`
	WindowTitle string `json:"window_title,omitempty"`
	Project     string `json:"project,omitempty"`
	Branch      string `json:"branch,omitempty"`
	MusicArtist string `json:"music_artist,omitempty"`
	MusicTrack  string `json:"music_track,omitempty"`
	Watching    string `json:"watching,omitempty"`
}

// HasMusic returns true if music is currently playing.
func (c Context) HasMusic() bool {
	return c.MusicTrack != ""
}

// Music returns "Artist - Track" if playing, empty string otherwise.
func (c Context) Music() string {
	if !c.HasMusic() {
		return ""
	}
	if c.MusicArtist != "" {
		return c.MusicArtist + " - " + c.MusicTrack
	}
	return c.MusicTrack
}

// HasWatching returns true if the user is watching something.
func (c Context) HasWatching() bool {
	return c.Watching != ""
}

// watchingSuffixes are window title suffixes for known streaming services.
var watchingSuffixes = []string{
	" - YouTube",
	" - Netflix",
	" - Twitch",
	" - Disney+",
	" - Prime Video",
	" _ Bilibili",
	" - Bilibili",
}

// detectWatching extracts what the user is watching from the window title.
// Returns the video/show title if a known streaming service is detected.
func detectWatching(windowTitle string) string {
	if windowTitle == "" {
		return ""
	}
	for _, suffix := range watchingSuffixes {
		if strings.HasSuffix(windowTitle, suffix) {
			title := strings.TrimSuffix(windowTitle, suffix)
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}
	return ""
}

// Detect gathers all available context from the current environment.
// It never returns an error — missing signals are silently skipped.
func Detect() Context {
	ctx := Context{}

	// Platform-specific: app + window title
	ctx.App, ctx.WindowTitle = detectApp()

	// Watching detection from window title
	ctx.Watching = detectWatching(ctx.WindowTitle)

	// Git info (cross-platform)
	ctx.Project, ctx.Branch = detectGit()

	// Platform-specific: music
	ctx.MusicArtist, ctx.MusicTrack = detectMusic()

	return ctx
}
