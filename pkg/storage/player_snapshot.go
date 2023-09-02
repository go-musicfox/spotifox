package storage

import (
	"time"

	"github.com/go-musicfox/spotifox/pkg/constants"
	"github.com/zmb3/spotify/v2"
)

type PlayerSnapshot struct {
	CurSongIndex     int                  `json:"cur_song_index"`
	Playlist         []*spotify.FullTrack `json:"playlist"`
	PlaylistUpdateAt time.Time            `json:"playlist_update_at"`
}

func (p PlayerSnapshot) GetDbName() string {
	return constants.AppDBName
}

func (p PlayerSnapshot) GetTableName() string {
	return "default_bucket"
}

func (p PlayerSnapshot) GetKey() string {
	return "playlist_snapshot"
}
