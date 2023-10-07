package ui

import (
	"context"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
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
	return "artist_song_" + string(m.artistId)
}

func (m *ArtistSongMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistSongMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.CheckAuthSession() == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		var country = "ES"
		if m.spotifox.user.Country != "" {
			country = m.spotifox.user.Country
		}
		res, err := m.spotifox.spotifyClient.GetArtistsTopTracks(context.Background(), m.artistId, country)
		if catched, page := m.spotifox.HandleResCode(utils.CheckSpotifyErr(err), EnterMenuCallback(main)); catched {
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "get artist's songs failed"))
		}

		m.songs = res
		m.menus = utils.MenuItemsFromSongs(m.songs)

		return true, nil
	}
}

func (m *ArtistSongMenu) Songs() []spotify.FullTrack {
	return m.songs
}
