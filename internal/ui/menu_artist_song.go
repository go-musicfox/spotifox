package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/zmb3/spotify/v2"
)

type ArtistSongMenu struct {
	baseMenu
	menus    []model.MenuItem
	songs    []spotify.FullTrack
	artistId spotify.ID
}

func NewArtistSongMenu(base baseMenu, artistId spotify.ID) *ArtistSongMenu {
	return &ArtistSongMenu{
		baseMenu: base,
		artistId: artistId,
	}
}

func (m *ArtistSongMenu) IsSearchable() bool {
	return true
}

func (m *ArtistSongMenu) IsPlayable() bool {
	return true
}

func (m *ArtistSongMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_song_%d", m.artistId)
}

func (m *ArtistSongMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistSongMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		// artistSongService := service.ArtistTopSongService{Id: string(m.artistId)}
		// code, response := artistSongService.ArtistTopSong()
		// codeType := utils.CheckCode(code)
		// if codeType != utils.Success {
		// 	return false, nil
		// }
		// m.songs = utils.GetSongsOfArtist(response)
		// m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *ArtistSongMenu) Songs() []spotify.FullTrack {
	return m.songs
}
