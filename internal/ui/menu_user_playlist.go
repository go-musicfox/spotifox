package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type UserPlaylistMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []spotify.SimplePlaylist
	userId    string
	offset    int
	limit     int
	total     int
}

const CurUser = "me"

func NewUserPlaylistMenu(base baseMenu, userId string) *UserPlaylistMenu {
	return &UserPlaylistMenu{
		baseMenu: base,
		userId:   userId,
		limit:    50,
	}
}

func (m *UserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *UserPlaylistMenu) GetMenuKey() string {
	return "user_playlist_" + m.userId
}

func (m *UserPlaylistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *UserPlaylistMenu) Playlists() []spotify.SimplePlaylist {
	return m.playlists
}

func (m *UserPlaylistMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.playlists) < index {
		return nil
	}
	return NewPlaylistDetailMenu(m.baseMenu, m.playlists[index].ID)
}

func (m *UserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		var (
			res *spotify.SimplePlaylistPage
			err error
		)
		if m.userId == CurUser {
			res, err = m.spotifox.spotifyClient.CurrentUsersPlaylists(context.Background(), spotify.Limit(m.limit))
		} else {
			res, err = m.spotifox.spotifyClient.GetPlaylistsForUser(context.Background(), m.userId, spotify.Limit(m.limit))
		}

		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), EnterMenuCallback(main)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlists failed"))
		}
		m.total = res.Total

		m.playlists = res.Playlists
		m.menus = utils.MenuItemsFromPlaylists(m.playlists)

		return true, nil
	}
}

func (m *UserPlaylistMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += m.limit
		var (
			res *spotify.SimplePlaylistPage
			err error
		)
		if m.userId == CurUser {
			res, err = m.spotifox.spotifyClient.CurrentUsersPlaylists(context.Background(), spotify.Limit(m.limit), spotify.Offset(m.offset))
		} else {
			res, err = m.spotifox.spotifyClient.GetPlaylistsForUser(context.Background(), m.userId, spotify.Limit(m.limit), spotify.Offset(m.offset))
		}

		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), BottomOutHookCallback(main, m)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlists failed"))
		}

		m.playlists = append(m.playlists, res.Playlists...)
		m.menus = utils.MenuItemsFromPlaylists(m.playlists)

		return true, nil
	}
}
