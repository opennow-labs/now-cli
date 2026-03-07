package detect

import "testing"

func TestClassifyMedia(t *testing.T) {
	tests := []struct {
		name string
		info *MediaInfo
		want MediaResult
	}{
		{
			name: "nil info",
			info: nil,
			want: MediaResult{},
		},
		{
			name: "not playing",
			info: &MediaInfo{Title: "Song", Artist: "Artist", SourceID: "com.spotify.client", IsPlaying: false},
			want: MediaResult{},
		},
		{
			name: "empty title",
			info: &MediaInfo{Title: "", Artist: "Artist", SourceID: "com.spotify.client", IsPlaying: true},
			want: MediaResult{},
		},
		{
			name: "spotify playing",
			info: &MediaInfo{Title: "Bohemian Rhapsody", Artist: "Queen", SourceID: "com.spotify.client", IsPlaying: true},
			want: MediaResult{MusicArtist: "Queen", MusicTrack: "Bohemian Rhapsody"},
		},
		{
			name: "apple music playing",
			info: &MediaInfo{Title: "Yesterday", Artist: "The Beatles", SourceID: "com.apple.Music", IsPlaying: true},
			want: MediaResult{MusicArtist: "The Beatles", MusicTrack: "Yesterday"},
		},
		{
			name: "linux spotify",
			info: &MediaInfo{Title: "Song", Artist: "Band", SourceID: "spotify", IsPlaying: true},
			want: MediaResult{MusicArtist: "Band", MusicTrack: "Song"},
		},
		{
			name: "vlc playing video",
			info: &MediaInfo{Title: "movie.mkv", SourceID: "org.videolan.vlc", IsPlaying: true},
			want: MediaResult{Watching: "movie.mkv"},
		},
		{
			name: "iina playing video",
			info: &MediaInfo{Title: "Episode 1", Artist: "Show", SourceID: "com.colliderli.iina", IsPlaying: true},
			want: MediaResult{Watching: "Show - Episode 1"},
		},
		{
			name: "linux mpv",
			info: &MediaInfo{Title: "video.mp4", SourceID: "mpv", IsPlaying: true},
			want: MediaResult{Watching: "video.mp4"},
		},
		{
			name: "chrome with album - music",
			info: &MediaInfo{Title: "Song", Artist: "Artist", Album: "Album", SourceID: "com.google.Chrome", IsPlaying: true},
			want: MediaResult{MusicArtist: "Artist", MusicTrack: "Song"},
		},
		{
			name: "chrome without album - watching",
			info: &MediaInfo{Title: "Funny Video", Artist: "Channel", SourceID: "com.google.Chrome", IsPlaying: true},
			want: MediaResult{Watching: "Channel - Funny Video"},
		},
		{
			name: "unknown source with album - music",
			info: &MediaInfo{Title: "Track", Artist: "Artist", Album: "Album", SourceID: "com.unknown.app", IsPlaying: true},
			want: MediaResult{MusicArtist: "Artist", MusicTrack: "Track"},
		},
		{
			name: "unknown source without album - watching",
			info: &MediaInfo{Title: "Content", SourceID: "com.unknown.app", IsPlaying: true},
			want: MediaResult{Watching: "Content"},
		},
		{
			name: "windows spotify keyword",
			info: &MediaInfo{Title: "Song", Artist: "Band", SourceID: "Spotify.exe", IsPlaying: true},
			want: MediaResult{MusicArtist: "Band", MusicTrack: "Song"},
		},
		{
			name: "windows vlc keyword",
			info: &MediaInfo{Title: "video", SourceID: "VideoLAN.VLC", IsPlaying: true},
			want: MediaResult{Watching: "video"},
		},
		{
			name: "windows zune music",
			info: &MediaInfo{Title: "Song", Artist: "Band", SourceID: "Microsoft.ZuneMusic_8wekyb3d8bbwe!Microsoft.ZuneMusic", IsPlaying: true},
			want: MediaResult{MusicArtist: "Band", MusicTrack: "Song"},
		},
		{
			name: "case insensitive source id",
			info: &MediaInfo{Title: "Song", Artist: "X", SourceID: "COM.SPOTIFY.CLIENT", IsPlaying: true},
			want: MediaResult{MusicArtist: "X", MusicTrack: "Song"},
		},
		{
			name: "apple tv - video",
			info: &MediaInfo{Title: "Movie", SourceID: "com.apple.TV", IsPlaying: true},
			want: MediaResult{Watching: "Movie"},
		},
		{
			name: "podcasts - music category",
			info: &MediaInfo{Title: "Episode 42", Artist: "Podcast", SourceID: "com.apple.podcasts", IsPlaying: true},
			want: MediaResult{MusicArtist: "Podcast", MusicTrack: "Episode 42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyMedia(tt.info)
			if got != tt.want {
				t.Errorf("ClassifyMedia() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
