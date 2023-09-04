package utils

import (
	"github.com/zmb3/spotify/v2"
)

func CompareSong(s1, s2 spotify.FullTrack) bool {
	if s1.ID == "" || s2.ID == "" {
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

type LyricsResp struct {
	Error    bool    `json:"error"`
	SyncType string  `json:"syncType"`
	Lines    []Lines `json:"lines"`
}
type Lines struct {
	StartTimeMs string        `json:"startTimeMs"`
	Words       string        `json:"words"`
	Syllables   []interface{} `json:"syllables"`
	EndTimeMs   string        `json:"endTimeMs"`
}
