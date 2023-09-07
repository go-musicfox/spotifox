package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/zmb3/spotify/v2"
)

type ArtistsOfSongMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
	song     spotify.FullTrack
}

func NewArtistsOfSongMenu(base baseMenu, song spotify.FullTrack) *ArtistsOfSongMenu {
	artistsMenu := &ArtistsOfSongMenu{
		song: song,
	}
	var subTitle = "「" + song.Name + "」所属歌手"
	for _, artist := range song.Artists {
		artistsMenu.menus = append(artistsMenu.menus, model.MenuItem{Title: artist.Name, Subtitle: subTitle})
		artistsMenu.menuList = append(artistsMenu.menuList, NewArtistDetailMenu(base, artist.ID, artist.Name))
	}

	return artistsMenu
}

func (m *ArtistsOfSongMenu) GetMenuKey() string {
	return "artist_of_song"
}

func (m *ArtistsOfSongMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistsOfSongMenu) Artists() []spotify.SimpleArtist {
	return m.song.Artists
}

func (m *ArtistsOfSongMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
