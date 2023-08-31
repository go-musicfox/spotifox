package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/structs"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type ArtistAlbumMenu struct {
	baseMenu
	menus    []model.MenuItem
	albums   []structs.Album
	artistId int64
}

func NewArtistAlbumMenu(base baseMenu, artistId int64) *ArtistAlbumMenu {
	return &ArtistAlbumMenu{
		baseMenu: base,
		artistId: artistId,
	}
}

func (m *ArtistAlbumMenu) IsSearchable() bool {
	return true
}

func (m *ArtistAlbumMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_album_%d", m.artistId)
}

func (m *ArtistAlbumMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistAlbumMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index].Id)
}

func (m *ArtistAlbumMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		artistAlbumService := service.ArtistAlbumService{
			ID:     strconv.FormatInt(m.artistId, 10),
			Offset: "0",
			Limit:  "50",
		}
		code, response := artistAlbumService.ArtistAlbum()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		m.albums = utils.GetArtistHotAlbums(response)
		m.menus = utils.GetViewFromAlbums(m.albums)

		return true, nil
	}
}

func (m *ArtistAlbumMenu) Albums() []structs.Album {
	return m.albums
}
