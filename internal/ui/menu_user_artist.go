package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type UserArtistMenu struct {
	baseMenu
	menus   []model.MenuItem
	artists []spotify.SimpleArtist
	offset  int
	limit   int
	total   int
}

func NewUserArtistMenu(base baseMenu) *UserArtistMenu {
	return &UserArtistMenu{
		baseMenu: base,
		limit:    50,
	}
}

func (m *UserArtistMenu) IsSearchable() bool {
	return true
}

func (m *UserArtistMenu) GetMenuKey() string {
	return "cur_user_artist"
}

func (m *UserArtistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *UserArtistMenu) Artists() []spotify.SimpleArtist {
	return m.artists
}

func (m *UserArtistMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.artists) < index {
		return nil
	}
	return NewArtistDetailMenu(m.baseMenu, m.artists[index].ID, m.artists[index].Name)
}

func (m *UserArtistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		res, err := m.spotifox.spotifyClient.CurrentUsersFollowedArtists(context.Background(), spotify.Limit(m.limit))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get artists failed"))
		}
		m.total = res.Total

		var artists []spotify.SimpleArtist
		for _, artist := range res.Artists {
			artists = append(artists, artist.SimpleArtist)
		}
		m.artists = artists
		m.menus = utils.MenuItemsFromArtists(m.artists)

		return true, nil
	}
}

func (m *UserArtistMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += m.limit
		res, err := m.spotifox.spotifyClient.CurrentUsersFollowedArtists(context.Background(), spotify.Limit(m.limit), spotify.Offset(m.offset))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get artists failed"))
		}

		for _, artist := range res.Artists {
			m.artists = append(m.artists, artist.SimpleArtist)
		}
		m.menus = utils.MenuItemsFromArtists(m.artists)

		return true, nil
	}
}
