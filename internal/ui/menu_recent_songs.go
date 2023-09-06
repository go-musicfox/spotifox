package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/zmb3/spotify/v2"
)

type RecentSongsMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []spotify.FullTrack
}

func NewRecentSongsMenu(base baseMenu) *RecentSongsMenu {
	return &RecentSongsMenu{
		baseMenu: base,
	}
}

func (m *RecentSongsMenu) IsSearchable() bool {
	return true
}

func (m *RecentSongsMenu) IsPlayable() bool {
	return true
}

func (m *RecentSongsMenu) GetMenuKey() string {
	return "recent_songs"
}

func (m *RecentSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *RecentSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// if m.spotifox.CheckAuthSession() == utils.NeedLogin {
		// 	page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
		// 	return false, page
		// }

		// recentSongService := service.RecordRecentSongsService{}
		// code, response := recentSongService.RecordRecentSongs()
		// codeType := utils.CheckCode(code)
		// if codeType == utils.NeedLogin {
		// 	page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
		// 	return false, page
		// } else if codeType != utils.Success {
		// 	return false, nil
		// }
		// m.songs = utils.GetRecentSongs(response)
		// m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *RecentSongsMenu) Songs() []spotify.FullTrack {
	return m.songs
}
