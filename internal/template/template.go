package template

import (
	"strings"

	"github.com/nownow-labs/nownow/internal/detect"
)

// Render takes a template string and a Context, returning the rendered status.
// Supported placeholders: {app}, {music}, {music.artist}, {music.track}, {title}, {emoji}, {watching}
// Empty placeholders are removed, and separators around them are cleaned up.
func Render(tmpl string, ctx detect.Context, emoji string) string {
	replacements := map[string]string{
		"{app}":          ctx.App,
		"{title}":        ctx.WindowTitle,
		"{music}":        ctx.Music(),
		"{music.artist}": ctx.MusicArtist,
		"{music.track}":  ctx.MusicTrack,
		"{watching}":     ctx.Watching,
		"{emoji}":        emoji,
	}

	result := tmpl
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Clean up artifacts from empty placeholders
	result = cleanSeparators(result)

	return strings.TrimSpace(result)
}

// cleanSeparators removes leftover separators and brackets from empty substitutions.
func cleanSeparators(s string) string {
	// Remove empty parentheses/brackets
	s = strings.ReplaceAll(s, "()", "")
	s = strings.ReplaceAll(s, "[]", "")

	// Collapse multiple dots/middots with spaces
	for strings.Contains(s, "· ·") {
		s = strings.ReplaceAll(s, "· ·", "·")
	}
	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	// Remove trailing/leading separators
	s = strings.TrimRight(s, " ·")
	s = strings.TrimLeft(s, " ·")

	return s
}
