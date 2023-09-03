package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/structs"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type ArtistSongMenu struct {
	baseMenu
	menus    []model.MenuItem
	songs    []structs.Song
	artistId int64
}

func NewArtistSongMenu(base baseMenu, artistId int64) *ArtistSongMenu {
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

		artistSongService := service.ArtistTopSongService{Id: strconv.FormatInt(m.artistId, 10)}
		code, response := artistSongService.ArtistTopSong()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}
		m.songs = utils.GetSongsOfArtist(response)
		m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *ArtistSongMenu) Songs() []structs.Song {
	return m.songs
}
