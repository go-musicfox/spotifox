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
		if m.spotifox.spotifyClient == nil || utils.CheckUserInfo(m.spotifox.user) == utils.NeedLogin {
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

		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlists failed"))
		}
		m.total = res.Total

		m.playlists = res.Playlists
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			var owner string
			if playlist.Owner.DisplayName != "" {
				owner = "[" + playlist.Owner.DisplayName + "]"
			}
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name), Subtitle: utils.ReplaceSpecialStr(owner)})
		}
		m.menus = menus

		return true, nil
	}
}

func (m *UserPlaylistMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset += len(m.menus)

		var (
			res *spotify.SimplePlaylistPage
			err error
		)
		if m.userId == CurUser {
			res, err = m.spotifox.spotifyClient.CurrentUsersPlaylists(context.Background(), spotify.Limit(m.limit))
		} else {
			res, err = m.spotifox.spotifyClient.GetPlaylistsForUser(context.Background(), m.userId, spotify.Limit(m.limit))
		}

		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlists failed"))
		}

		m.playlists = append(m.playlists, res.Playlists...)
		var menus []model.MenuItem
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name), Subtitle: utils.ReplaceSpecialStr(playlist.Owner.DisplayName)})
		}
		m.menus = menus

		return true, nil
	}
}
