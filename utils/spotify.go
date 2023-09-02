package utils

import (
	"github.com/zmb3/spotify/v2"
)

func CompareSong(s1, s2 *spotify.FullTrack) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	return s1.ID == s2.ID
}

func ArtistNamesOfSong(s *spotify.FullTrack) []string {
	var names []string
	for _, a := range s.Artists {
		names = append(names, a.Name)
	}
	return names
}
