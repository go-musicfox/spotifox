package state_handler

import (
	"time"

	"github.com/go-musicfox/spotifox/internal/player"
)

type PlayingInfo struct {
	TotalDuration  time.Duration
	PassedDuration time.Duration
	State          player.State
	Volume         int
	TrackID        string
	PicUrl         string
	Name           string
	Artist         string
	Album          string
}
