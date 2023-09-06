package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type AlbumDetailMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []spotify.FullTrack
	album spotify.SimpleAlbum
}

func NewAlbumDetailMenu(base baseMenu, album spotify.SimpleAlbum) *AlbumDetailMenu {
	return &AlbumDetailMenu{
		baseMenu: base,
		album:    album,
	}
}

func (m *AlbumDetailMenu) IsSearchable() bool {
	return true
}

func (m *AlbumDetailMenu) IsPlayable() bool {
	return true
}

func (m *AlbumDetailMenu) GetMenuKey() string {
	return "album_detail_" + string(m.album.ID)
}

func (m *AlbumDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AlbumDetailMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		res, err := m.spotifox.spotifyClient.GetAlbumTracks(context.Background(), m.album.ID, spotify.Limit(50))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get album's songs failed"))
		}

		var songs []spotify.FullTrack
		for _, song := range res.Tracks {
			songs = append(songs, spotify.FullTrack{
				Album:       m.album,
				SimpleTrack: song,
			})
		}
		m.songs = songs
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *AlbumDetailMenu) Songs() []spotify.FullTrack {
	return m.songs
}
