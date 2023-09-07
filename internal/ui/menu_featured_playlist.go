package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type FeaturedPlaylistMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []spotify.SimplePlaylist
	offset    int
	limit     int
	total     int
}

func NewFeaturedPlaylistMenu(base baseMenu) *FeaturedPlaylistMenu {
	return &FeaturedPlaylistMenu{
		baseMenu: base,
		limit:    50,
	}
}

func (m *FeaturedPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *FeaturedPlaylistMenu) GetMenuKey() string {
	return "featured_playlist"
}

func (m *FeaturedPlaylistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *FeaturedPlaylistMenu) Playlists() []spotify.SimplePlaylist {
	return m.playlists
}

func (m *FeaturedPlaylistMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.playlists) < index {
		return nil
	}
	return NewPlaylistDetailMenu(m.baseMenu, m.playlists[index].ID)
}

func (m *FeaturedPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		msg, res, err := m.spotifox.spotifyClient.FeaturedPlaylists(context.Background(), spotify.Limit(m.limit))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get featured playlists failed"))
		}
		m.total = res.Total

		tips := model.NewMenuTips(main, main.MenuTitle())
		tips.DisplayTips("「" + msg + "」")

		m.playlists = res.Playlists
		m.menus = utils.MenuItemsFromPlaylists(m.playlists)

		return true, nil
	}
}

func (m *FeaturedPlaylistMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += m.limit
		_, res, err := m.spotifox.spotifyClient.FeaturedPlaylists(context.Background(), spotify.Limit(m.limit), spotify.Offset(m.offset))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get featured playlists failed"))
		}

		m.playlists = append(m.playlists, res.Playlists...)
		m.menus = utils.MenuItemsFromPlaylists(m.playlists)

		return true, nil
	}
}
