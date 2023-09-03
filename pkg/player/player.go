package player

import (
	"time"

	"github.com/go-musicfox/spotifox/pkg/configs"
	"github.com/go-musicfox/spotifox/pkg/constants"
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
	switch registry.PlayerEngine {
	case constants.BeepPlayer, constants.OsxPlayer:
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
	PmListLoop:   "列表",
	PmOrder:      "顺序",
	PmSingleLoop: "单曲",
	PmRandom:     "随机",
}

func ModeName(mode Mode) string {
	if name, ok := modeNames[mode]; ok {
		return name
	}
	return "未知"
}
