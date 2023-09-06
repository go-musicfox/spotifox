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
