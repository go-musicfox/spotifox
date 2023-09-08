package player

import (
	"time"

	"github.com/arcspace/go-arc-sdk/apis/arc"
	"github.com/zmb3/spotify/v2"
)

type SongType uint8

const (
	Mp3 SongType = iota
	Ogg
	Aac
)

type MediaAsset struct {
	arc.MediaAsset
	SongInfo spotify.FullTrack
}

func (m MediaAsset) Duration() time.Duration {
	return m.SongInfo.TimeDuration()
}

func (m MediaAsset) SongType() SongType {
	switch m.MediaType() {
	case "audio/mpeg":
		return Mp3
	case "audio/ogg":
		return Ogg
	case "audio/aac":
		return Aac
	default:
		return Ogg
	}
}
