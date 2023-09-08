package storage

import (
	"time"

	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/zmb3/spotify/v2"
)

type PlayerSnapshot struct {
	CurSongIndex     int                 `json:"cur_song_index"`
	Playlist         []spotify.FullTrack `json:"playlist"`
	PlaylistUpdateAt time.Time           `json:"playlist_update_at"`
	IsCurSongLiked   bool                `json:"is_cur_song_liked"`
}

func (p PlayerSnapshot) GetDbName() string {
	return types.AppDBName
}

func (p PlayerSnapshot) GetTableName() string {
	return "default_bucket"
}

func (p PlayerSnapshot) GetKey() string {
	return "playlist_snapshot"
}
