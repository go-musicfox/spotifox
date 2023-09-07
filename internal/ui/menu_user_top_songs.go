package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type UserTopSongsMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []spotify.FullTrack

	limit  int
	offset int
	total  int
}

func NewUserTopSongsMenu(base baseMenu) *UserTopSongsMenu {
	return &UserTopSongsMenu{
		baseMenu: base,
	}
}

func (m *UserTopSongsMenu) IsSearchable() bool {
	return true
}

func (m *UserTopSongsMenu) IsPlayable() bool {
	return true
}

func (m *UserTopSongsMenu) GetMenuKey() string {
	return "user_top_song"
}

func (m *UserTopSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *UserTopSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		res, err := m.spotifox.spotifyClient.CurrentUsersTopTracks(context.Background(), spotify.Limit(m.limit))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get current user top tracks failed"))
		}
		m.total = res.Total

		m.songs = res.Tracks
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *UserTopSongsMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += m.limit
		res, err := m.spotifox.spotifyClient.CurrentUsersTopTracks(context.Background(), spotify.Limit(m.limit), spotify.Offset(m.offset))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get current user tracks failed"))
		}
		m.songs = append(m.songs, res.Tracks...)
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *UserTopSongsMenu) Songs() []spotify.FullTrack {
	return m.songs
}
