package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/zmb3/spotify/v2"
)

type AlbumDetailMenu struct {
	baseMenu
	menus   []model.MenuItem
	songs   []spotify.FullTrack
	albumId spotify.ID
}

func NewAlbumDetailMenu(base baseMenu, albumId spotify.ID) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		baseMenu: base,
		albumId:  albumId,
	}
}

func (m *AlbumDetailMenu) IsSearchable() bool {
	return true
}

func (m *AlbumDetailMenu) IsPlayable() bool {
	return true
}

func (m *AlbumDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("album_detail_%d", m.albumId)
}

func (m *AlbumDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		// albumService := service.AlbumService{
		// 	ID: string(m.albumId),
		// }
		// code, response := albumService.Album()
		// codeType := utils.CheckCode(code)
		// if codeType == utils.NeedLogin {
		// 	page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
		// 	return false, page
		// } else if codeType != utils.Success {
		// 	return false, nil
		// }

		// m.songs = utils.GetSongsOfAlbum(response)
		// m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *AlbumDetailMenu) Songs() []spotify.FullTrack {
	return m.songs
}
