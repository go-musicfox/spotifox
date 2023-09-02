package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/structs"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"
)

// Menu menu interface
type Menu interface {
	model.Menu

	// IsPlayable 当前菜单是否可播放？
	IsPlayable() bool

	// IsLocatable 当前菜单是否支持播放自动定位
	IsLocatable() bool
}

type SongsMenu interface {
	Menu
	Songs() []*spotify.FullTrack
}

type PlaylistsMenu interface {
	Menu
	Playlists() []spotify.SimplePlaylist
}

type AlbumsMenu interface {
	Menu
	Albums() []structs.Album
}

type ArtistsMenu interface {
	Menu
	Artists() []structs.Artist
}

type baseMenu struct {
	model.DefaultMenu
	spotifox *Spotifox
}

func newBaseMenu(spotifox *Spotifox) baseMenu {
	return baseMenu{
		spotifox: spotifox,
	}
}

func (e *baseMenu) IsPlayable() bool {
	return false
}

func (e *baseMenu) IsLocatable() bool {
	return true
}

func (e *baseMenu) handleFetchErr(err error) (bool, model.Page) {
	utils.Logger().Printf("[ERROR] err: %+v", err)
	model.NewMenuTips(e.spotifox.MustMain(), nil).DisplayTips("Err:" + err.Error())
	return false, nil
}
