package utils

import (
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/zmb3/spotify/v2"
)

func MenuItemsFromSongs(songs []spotify.FullTrack) []model.MenuItem {
	var menus []model.MenuItem
	for _, song := range songs {
		var artists []string
		for _, a := range song.Artists {
			artists = append(artists, a.Name)
		}
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(song.Name), Subtitle: ReplaceSpecialStr(strings.Join(artists, ","))})
	}
	return menus
}

func MenuItemsFromAlbums(albums []spotify.SimpleAlbum) []model.MenuItem {
	var menus []model.MenuItem
	for _, album := range albums {
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(album.Name), Subtitle: "[" + ReplaceSpecialStr(ArtistNameStrOfAlbum(&album)) + "]"})
	}
	return menus
}

func MenuItemsFromArtists(artists []spotify.SimpleArtist) []model.MenuItem {
	var menus []model.MenuItem
	for _, artist := range artists {
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(artist.Name)})
	}
	return menus
}

func MenuItemsFromPlaylists(playlists []spotify.SimplePlaylist) []model.MenuItem {
	var menus []model.MenuItem
	for _, playlist := range playlists {
		var owner string
		if playlist.Owner.DisplayName != "" {
			owner = "[" + playlist.Owner.DisplayName + "]"
		}
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(playlist.Name), Subtitle: ReplaceSpecialStr(owner)})
	}
	return menus
}
