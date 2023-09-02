package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"
)

const CurPlaylistKey = "cur_playlist"

type CurPlaylist struct {
	baseMenu
	menus []model.MenuItem
	songs []*spotify.FullTrack
}

func NewCurPlaylist(base baseMenu, songs []*spotify.FullTrack) *CurPlaylist {
	return &CurPlaylist{
		baseMenu: base,
		songs:    songs,
		menus:    utils.MenuItemsFromSongs(songs),
	}
}

func (m *CurPlaylist) IsSearchable() bool {
	return true
}

func (m *CurPlaylist) IsPlayable() bool {
	return true
}

func (m *CurPlaylist) GetMenuKey() string {
	return CurPlaylistKey
}

func (m *CurPlaylist) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *CurPlaylist) Songs() []*spotify.FullTrack {
	return m.songs
}

func (m *CurPlaylist) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.player.playingMenu == nil || m.spotifox.player.playingMenu.GetMenuKey() == CurPlaylistKey {
			return true, nil
		}
		hook := m.spotifox.player.playingMenu.BottomOutHook()
		if hook == nil {
			return true, nil
		}
		res, page := hook(main)
		m.songs = m.spotifox.player.playlist
		m.menus = utils.MenuItemsFromSongs(m.songs)
		return res, page
	}
}
