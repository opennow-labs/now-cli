package detect

// Context holds the detected environment context.
type Context struct {
	App         string `json:"app,omitempty"`
	WindowTitle string `json:"window_title,omitempty"`
	Project     string `json:"project,omitempty"`
	Branch      string `json:"branch,omitempty"`
	MusicArtist string `json:"music_artist,omitempty"`
	MusicTrack  string `json:"music_track,omitempty"`
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

// Detect gathers all available context from the current environment.
// It never returns an error — missing signals are silently skipped.
func Detect() Context {
	ctx := Context{}

	// Platform-specific: app + window title
	ctx.App, ctx.WindowTitle = detectApp()

	// Git info (cross-platform)
	ctx.Project, ctx.Branch = detectGit()

	// Platform-specific: music
	ctx.MusicArtist, ctx.MusicTrack = detectMusic()

	return ctx
}
