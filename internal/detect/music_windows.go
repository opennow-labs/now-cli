//go:build windows

package detect

// detectMusic is a stub on Windows — no reliable local IPC for music players.
// Users can extend this in the future with Spotify Web API integration.
func detectMusic() (artist, track string) {
	return "", ""
}
