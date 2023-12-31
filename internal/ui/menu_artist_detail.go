package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/utils/locale"
	"github.com/zmb3/spotify/v2"
)

type ArtistDetailMenu struct {
	baseMenu
	menus    []model.MenuItem
	artistId spotify.ID
}

func NewArtistDetailMenu(base baseMenu, artistId spotify.ID, artistName string) *ArtistDetailMenu {
	artistMenu := &ArtistDetailMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: locale.MustT("artist_top_track"), Subtitle: artistName},
			{Title: locale.MustT("artist_album"), Subtitle: artistName},
		},
		artistId: artistId,
	}

	return artistMenu
}

func (m *ArtistDetailMenu) GetMenuKey() string {
	return "artist_detail_" + string(m.artistId)
}

func (m *ArtistDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistDetailMenu) SubMenu(_ *model.App, index int) model.Menu {
	switch index {
	case 0:
		return NewArtistSongMenu(m.baseMenu, m.artistId)
	case 1:
		return NewArtistAlbumMenu(m.baseMenu, m.artistId)
	}

	return nil
}
