package player

import (
	"time"

	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils/locale"
)

type Player interface {
	Play(music MediaAsset)
	CurMusic() MediaAsset
	Paused()
	Resume()
	Stop()
	Toggle()
	Seek(duration time.Duration)
	PassedTime() time.Duration
	TimeChan() <-chan time.Duration
	State() State
	StateChan() <-chan State
	Volume() int
	SetVolume(volume int)
	UpVolume()
	DownVolume()
	Close()
}

func NewPlayerFromConfig() Player {
	registry := configs.ConfigRegistry
	var player Player
	switch registry.Player.Engine {
	case types.BeepPlayer, types.OsxPlayer:
		player = NewBeepPlayer()
	// case constants.OsxPlayer:
	// 	player = NewOsxPlayer()
	default:
		panic("unknown player engine")
	}

	return player
}

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)

// Mode play mode
type Mode uint8

const (
	PmListLoop Mode = iota + 1
	PmOrder
	PmSingleLoop
	PmRandom
)

var modeNames = map[Mode]string{
	PmListLoop:   locale.MustT("list_loop"),
	PmOrder:      locale.MustT("order"),
	PmSingleLoop: locale.MustT("single_loop"),
	PmRandom:     locale.MustT("random"),
}

func ModeName(mode Mode) string {
	if name, ok := modeNames[mode]; ok {
		return name
	}
	return locale.MustT("unknown")
}
