package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type LikedSongsMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []spotify.FullTrack

	limit  int
	offset int
	total  int
}

func NewLikedSongsMenu(base baseMenu) *LikedSongsMenu {
	return &LikedSongsMenu{
		baseMenu: base,
		limit:    50,
	}
}

func (m *LikedSongsMenu) IsSearchable() bool {
	return true
}

func (m *LikedSongsMenu) IsPlayable() bool {
	return true
}

func (m *LikedSongsMenu) GetMenuKey() string {
	return "liked_songs"
}

func (m *LikedSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *LikedSongsMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *LikedSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		res, err := m.spotifox.spotifyClient.CurrentUsersTracks(context.Background(), spotify.Limit(m.limit))
		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), EnterMenuCallback(main)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get current user tracks failed"))
		}
		m.total = res.Total

		var songs []spotify.FullTrack
		for i := range res.Tracks {
			songs = append(songs, res.Tracks[i].FullTrack)
		}
		m.songs = songs
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *LikedSongsMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += m.limit
		res, err := m.spotifox.spotifyClient.CurrentUsersTracks(context.Background(), spotify.Limit(m.limit), spotify.Offset(m.offset))
		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), BottomOutHookCallback(main, m)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get current user tracks failed"))
		}
		for i := range res.Tracks {
			m.songs = append(m.songs, res.Tracks[i].FullTrack)
		}
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *LikedSongsMenu) Songs() []spotify.FullTrack {
	return m.songs
}
