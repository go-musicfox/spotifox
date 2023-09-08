package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils/locale"
	"github.com/zmb3/spotify/v2"
)

type SearchTypeMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewSearchTypeMenu(base baseMenu) *SearchTypeMenu {
	typeMenu := &SearchTypeMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: locale.MustT("search_track")},
			{Title: locale.MustT("search_album")},
			{Title: locale.MustT("search_artist")},
			{Title: locale.MustT("search_playlist")},
		},
	}

	return typeMenu
}

func (m *SearchTypeMenu) GetMenuKey() string {
	return "search_type"
}

func (m *SearchTypeMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SearchTypeMenu) SubMenu(_ *model.App, index int) model.Menu {
	typeArr := []spotify.SearchType{
		spotify.SearchTypeTrack,
		spotify.SearchTypeAlbum,
		spotify.SearchTypeArtist,
		spotify.SearchTypePlaylist,
	}

	if index >= len(typeArr) {
		return nil
	}

	return NewSearchResultMenu(m.baseMenu, typeArr[index])
}
