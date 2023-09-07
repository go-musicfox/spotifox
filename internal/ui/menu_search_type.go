package ui

import (
	"github.com/anhoder/foxful-cli/model"
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
			{Title: "搜单曲"},
			{Title: "搜专辑"},
			{Title: "搜歌手"},
			{Title: "搜歌单"},
			// {Title: "搜插曲"},
			// {Title: "搜演出"},
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
		// spotify.SearchTypeEpisode,
		// spotify.SearchTypeShow,
	}

	if index >= len(typeArr) {
		return nil
	}

	return NewSearchResultMenu(m.baseMenu, typeArr[index])
}
