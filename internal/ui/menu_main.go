package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils/locale"
)

type MainMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewMainMenu(netease *Spotifox) *MainMenu {
	base := newBaseMenu(netease)
	mainMenu := &MainMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: locale.MustT("liked_tracks")},
			{Title: locale.MustT("followed_playlists")},
			{Title: locale.MustT("followed_artists")},
			{Title: locale.MustT("featured_playlist")},
			// {Title: locale.MustT("my_top_tracks")},
			{Title: locale.MustT("search")},
			{Title: "LastFM"},
			{Title: locale.MustT("help")},
			{Title: locale.MustT("check_update")},
		},
		menuList: []Menu{
			NewLikedSongsMenu(base),
			NewUserPlaylistMenu(base, CurUser),
			NewUserArtistMenu(base),
			NewFeaturedPlaylistMenu(base),
			// NewUserTopSongsMenu(base),
			NewSearchTypeMenu(base),
			NewLastfm(base),
			NewHelpMenu(base),
			NewCheckUpdateMenu(base),
		},
	}
	return mainMenu
}

func (m *MainMenu) FormatMenuItem(item *model.MenuItem) {
	if m.spotifox.user == nil {
		item.Subtitle = "[" + locale.MustT("no_login") + "]"
		return
	}
	if m.spotifox.user.DisplayName != "" {
		item.Subtitle = "[" + m.spotifox.user.DisplayName + "]"
		return
	}
	item.Subtitle = "[" + m.spotifox.user.Username + "]"
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	for i, menu := range m.menuList {
		menu.FormatMenuItem(&m.menus[i])
	}
	return m.menus
}

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
