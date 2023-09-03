package utils

import (
	"encoding/json"
	"io"
	"net/http"

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

func GetExternalLyrics(id spotify.ID) *LyricsResp {
	resp, err := http.Get("https://spotify-lyric-api.herokuapp.com/?trackid=" + string(id))
	if err != nil {
		Logger().Printf("Get song lyrics failed: %+v", err)
		return nil
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger().Printf("Get song lyrics failed: %+v", err)
		return nil
	}

	var lyrics LyricsResp
	if err = json.Unmarshal(data, &lyrics); err != nil {
		Logger().Printf("Get song lyrics failed: %+v", err)
		return nil
	}

	return &lyrics
}
