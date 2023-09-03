package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type PlaylistDetailMenu struct {
	baseMenu
	menus      []model.MenuItem
	songs      []spotify.FullTrack
	playlistId spotify.ID

	limit  int
	offset int
	total  int
}

func NewPlaylistDetailMenu(base baseMenu, playlistId spotify.ID) *PlaylistDetailMenu {
	return &PlaylistDetailMenu{
		baseMenu:   base,
		playlistId: playlistId,

		limit: 50,
	}
}

func (m *PlaylistDetailMenu) IsSearchable() bool {
	return true
}

func (m *PlaylistDetailMenu) IsPlayable() bool {
	return true
}

func (m *PlaylistDetailMenu) GetMenuKey() string {
	return "playlist_detail_" + string(m.playlistId)
}

func (m *PlaylistDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *PlaylistDetailMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *PlaylistDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		res, err := m.spotifox.spotifyClient.GetPlaylistItems(context.Background(), m.playlistId, spotify.Limit(m.limit))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlist items failed"))
		}
		m.total = res.Total

		for _, v := range res.Items {
			if v.Track.Track == nil {
				continue
			}
			m.songs = append(m.songs, *v.Track.Track)
		}
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *PlaylistDetailMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += len(m.menus)
		res, err := m.spotifox.spotifyClient.GetPlaylistItems(context.Background(), m.playlistId, spotify.Limit(m.limit))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlist items failed"))
		}
		for _, v := range res.Items {
			if v.Track.Track == nil {
				continue
			}
			m.songs = append(m.songs, *v.Track.Track)
		}
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *PlaylistDetailMenu) Songs() []spotify.FullTrack {
	return m.songs
}
