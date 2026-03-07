package detect

import "strings"

// MediaInfo holds raw now-playing data from platform-specific APIs.
type MediaInfo struct {
	Title     string
	Artist    string
	Album     string
	SourceID  string // bundleID (macOS), player name (Linux), appModelId (Windows)
	IsPlaying bool
}

// MediaResult holds classified media output.
type MediaResult struct {
	MusicArtist string
	MusicTrack  string
	Watching    string
}

// Known music app source IDs (lowercase for case-insensitive matching).
var musicSourceIDs = map[string]bool{
	// macOS bundle IDs
	"com.spotify.client":       true,
	"com.apple.music":          true,
	"com.tidal.desktop":        true,
	"com.amazon.music":         true,
	"com.deezer.deezer-desktop": true,
	"com.qobuz.qobuzdesktop":  true,
	"com.netease.163music":     true,
	"com.tencent.qqmusicmac":   true,
	"com.coppertino.vox":       true,
	"com.swinsian.swinsian":    true,
	"org.cogx.cog":             true,
	"com.foobar2000.mac":       true,
	"com.apple.podcasts":       true,
	// Linux player names
	"spotify":              true,
	"music":                true,
	"tidal-hifi":           true,
	"amazonmusic":          true,
	"netease-cloud-music":  true,
	// Windows appModelId keywords handled by containsAny below
}

// Known video app source IDs (lowercase).
var videoSourceIDs = map[string]bool{
	// macOS bundle IDs
	"org.videolan.vlc":           true,
	"com.colliderli.iina":        true,
	"io.mpv":                     true,
	"com.apple.quicktimeplayerx": true,
	"com.apple.tv":               true,
	"com.movist.movistpro":       true,
	"com.firecore.infuse":        true,
	"com.bilibili.bilibilipc":    true,
	// Linux player names
	"vlc":       true,
	"mpv":       true,
	"celluloid": true,
	"totem":     true,
}

// Windows appModelId substrings for music apps.
var windowsMusicKeywords = []string{
	"spotify", "zunemusic", "tidal", "deezer", "foobar2000",
}

// Windows appModelId substrings for video apps.
var windowsVideoKeywords = []string{
	"vlc", "mpv",
}

// ClassifyMedia determines whether MediaInfo represents music, video, or nothing.
func ClassifyMedia(info *MediaInfo) MediaResult {
	if info == nil || !info.IsPlaying || info.Title == "" {
		return MediaResult{}
	}

	sid := strings.ToLower(info.SourceID)

	// Check known music apps
	if musicSourceIDs[sid] || containsAny(sid, windowsMusicKeywords) {
		return MediaResult{
			MusicArtist: info.Artist,
			MusicTrack:  info.Title,
		}
	}

	// Check known video apps
	if videoSourceIDs[sid] || containsAny(sid, windowsVideoKeywords) {
		return MediaResult{
			Watching: formatWatching(info.Title, info.Artist),
		}
	}

	// Unknown source (browser, etc.): album non-empty → music, otherwise → watching
	if info.Album != "" {
		return MediaResult{
			MusicArtist: info.Artist,
			MusicTrack:  info.Title,
		}
	}
	return MediaResult{
		Watching: formatWatching(info.Title, info.Artist),
	}
}

func formatWatching(title, artist string) string {
	if artist != "" {
		return artist + " - " + title
	}
	return title
}

func containsAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
