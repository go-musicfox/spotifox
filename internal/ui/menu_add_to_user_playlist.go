package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type AddToUserPlaylistMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []spotify.SimplePlaylist
	song      spotify.FullTrack
	offset    int
	limit     int
	total     int
	action    bool // true for add, false for del
}

func NewAddToUserPlaylistMenu(base baseMenu, song spotify.FullTrack, action bool) *AddToUserPlaylistMenu {
	return &AddToUserPlaylistMenu{
		baseMenu: base,
		limit:    50,
		action:   action,
		song:     song,
	}
}

func (m *AddToUserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *AddToUserPlaylistMenu) GetMenuKey() string {
	return "add_to_user_playlist_" + m.spotifox.user.ID
}

func (m *AddToUserPlaylistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AddToUserPlaylistMenu) Playlists() []spotify.SimplePlaylist {
	return m.playlists
}

func (m *AddToUserPlaylistMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *AddToUserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 等于0，获取当前用户歌单
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		var (
			res *spotify.SimplePlaylistPage
			err error
		)
		res, err = m.spotifox.spotifyClient.CurrentUsersPlaylists(context.Background(), spotify.Limit(m.limit))
		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), EnterMenuCallback(main)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlists failed"))
		}
		m.total = res.Total

		var (
			menus     []model.MenuItem
			playlists []spotify.SimplePlaylist
		)
		for _, playlist := range res.Playlists {
			if playlist.Owner.ID != m.spotifox.user.ID {
				continue
			}
			playlists = append(playlists, playlist)
			var owner string
			if playlist.Owner.DisplayName != "" {
				owner = "[" + playlist.Owner.DisplayName + "]"
			}
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name), Subtitle: utils.ReplaceSpecialStr(owner)})
		}
		m.playlists = playlists
		m.menus = menus

		return true, nil
	}
}

func (m *AddToUserPlaylistMenu) BottomOutHook() model.Hook {
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
		res, err = m.spotifox.spotifyClient.CurrentUsersPlaylists(context.Background(), spotify.Limit(m.limit), spotify.Offset(m.offset))
		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), BottomOutHookCallback(main, m)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get playlists failed"))
		}

		for _, playlist := range res.Playlists {
			if playlist.Owner.ID != m.spotifox.user.ID {
				continue
			}
			m.playlists = append(m.playlists, playlist)
			var owner string
			if playlist.Owner.DisplayName != "" {
				owner = "[" + playlist.Owner.DisplayName + "]"
			}
			m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name), Subtitle: utils.ReplaceSpecialStr(owner)})
		}

		return true, nil
	}
}
