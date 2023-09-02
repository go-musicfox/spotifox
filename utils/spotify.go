package utils

import (
	"time"

	"github.com/zmb3/spotify/v2"
)

func CompareSong(s1, s2 spotify.PlaylistItem) bool {
	s1ID := IDOfSong(s1)
	s2ID := IDOfSong(s2)

	if s1ID == "" || s2ID == "" {
		return false
	}
	return s1ID == s2ID
}

func IDOfSong(s spotify.PlaylistItem) spotify.ID {
	if s.Track.Track != nil {
		return s.Track.Track.ID
	}
	if s.Track.Episode != nil {
		return s.Track.Episode.ID
	}
	return ""
}

func NameOfSong(s spotify.PlaylistItem) string {
	if s.Track.Track != nil {
		return s.Track.Track.Name
	}
	if s.Track.Episode != nil {
		return s.Track.Episode.Name
	}
	return ""
}

func ArtistsOfSong(s spotify.PlaylistItem) []spotify.SimpleArtist {
	if s.Track.Track != nil {
		return s.Track.Track.Artists
	}
	return nil
}

func ArtistNamesOfSong(s spotify.PlaylistItem) []string {
	if s.Track.Track == nil {
		return nil
	}
	var names []string
	for _, a := range s.Track.Track.Artists {
		names = append(names, a.Name)
	}
	return names
}

func AlbumNameOfSong(s spotify.PlaylistItem) string {
	if s.Track.Track != nil {
		return s.Track.Track.Album.Name
	}
	return ""
}

func DurationOfSong(s spotify.PlaylistItem) time.Duration {
	if s.Track.Track != nil {
		return s.Track.Track.TimeDuration()
	}
	if s.Track.Episode != nil {
		return time.Duration(s.Track.Episode.Duration_ms) * time.Millisecond
	}
	return 0
}
