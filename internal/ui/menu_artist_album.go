package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

const AllType spotify.AlbumType = spotify.AlbumTypeAlbum | spotify.AlbumTypeSingle | spotify.AlbumTypeAppearsOn | spotify.AlbumTypeCompilation

type ArtistAlbumMenu struct {
	baseMenu
	menus    []model.MenuItem
	albums   []spotify.SimpleAlbum
	artistId spotify.ID
	limit    int
	offset   int
	total    int
}

func NewArtistAlbumMenu(base baseMenu, artistId spotify.ID) *ArtistAlbumMenu {
	return &ArtistAlbumMenu{
		baseMenu: base,
		artistId: artistId,
		limit:    50,
	}
}

func (m *ArtistAlbumMenu) IsSearchable() bool {
	return true
}

func (m *ArtistAlbumMenu) GetMenuKey() string {
	return "artist_album_" + string(m.artistId)
}

func (m *ArtistAlbumMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistAlbumMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return NewAlbumDetailMenu(m.baseMenu, m.albums[index])
}

func (m *ArtistAlbumMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		res, err := m.spotifox.spotifyClient.GetArtistAlbums(context.Background(), m.artistId, []spotify.AlbumType{AllType}, spotify.Limit(m.limit))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get artist's album failed"))
		}
		m.total = res.Total

		m.albums = res.Albums
		var menus []model.MenuItem
		for _, album := range m.albums {
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: "[" + utils.ReplaceSpecialStr(utils.ArtistNameStrOfAlbum(&album)) + "]"})
		}
		m.menus = menus

		return true, nil
	}
}

func (m *ArtistAlbumMenu) Albums() []spotify.SimpleAlbum {
	return m.albums
}

func (m *ArtistAlbumMenu) BottomOutHook() model.Hook {
	if m.total <= m.limit+m.offset {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}

		m.offset += m.limit
		res, err := m.spotifox.spotifyClient.GetArtistAlbums(context.Background(), m.artistId, []spotify.AlbumType{AllType}, spotify.Limit(m.limit), spotify.Offset(m.offset))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get artist's album failed"))
		}

		m.albums = append(m.albums, res.Albums...)
		var menus []model.MenuItem
		for _, album := range m.albums {
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(album.Name), Subtitle: "[" + utils.ReplaceSpecialStr(utils.ArtistNameStrOfAlbum(&album)) + "]"})
		}
		m.menus = menus

		return true, nil
	}
}
