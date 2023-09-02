package utils

import (
	"fmt"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	ds "github.com/go-musicfox/spotifox/pkg/structs"
	"github.com/zmb3/spotify/v2"
)

func GetViewFromSongs(songs []ds.Song) []model.MenuItem {
	return nil
}

func MenuItemsFromSongs(songs []spotify.PlaylistItem) []model.MenuItem {
	var (
		menus    []model.MenuItem
		title    string
		subtitle string
	)
	for _, song := range songs {
		if song.Track.Track != nil {
			title = song.Track.Track.Name
			var artists []string
			for _, a := range song.Track.Track.Artists {
				artists = append(artists, a.Name)
			}
			subtitle = strings.Join(artists, ",")
		} else if song.Track.Episode != nil {
			title = song.Track.Episode.Name
			subtitle = song.Track.Episode.Description
		}
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(title), Subtitle: ReplaceSpecialStr(subtitle)})
	}

	return menus
}

// GetViewFromAlbums 从歌曲列表获取View
func GetViewFromAlbums(albums []ds.Album) []model.MenuItem {
	var menus []model.MenuItem
	for _, album := range albums {
		var artists []string
		for _, artist := range album.Artists {
			artists = append(artists, artist.Name)
		}
		artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(album.Name), Subtitle: ReplaceSpecialStr(artistsStr)})
	}

	return menus
}

// GetViewFromPlaylists 从歌单列表获取View
func GetViewFromPlaylists(playlists []ds.Playlist) []model.MenuItem {
	var menus []model.MenuItem
	for _, playlist := range playlists {
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(playlist.Name)})
	}

	return menus
}

// GetViewFromArtists 从歌手列表获取View
func GetViewFromArtists(artists []ds.Artist) []model.MenuItem {
	var menus []model.MenuItem
	for _, artist := range artists {
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(artist.Name)})
	}

	return menus
}

// GetViewFromUsers 用户列表获取View
func GetViewFromUsers(users []ds.User) []model.MenuItem {
	var menus []model.MenuItem
	//for _, user := range users {
	//	menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(user.Nickname)})
	//}

	return menus
}

// GetViewFromDjRadios DjRadio列表获取View
func GetViewFromDjRadios(radios []ds.DjRadio) []model.MenuItem {
	var menus []model.MenuItem
	for _, radio := range radios {
		var dj string
		//if radio.Dj.Nickname != "" {
		//	dj = fmt.Sprintf("[%s]", radio.Dj.Nickname)
		//}
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(radio.Name), Subtitle: ReplaceSpecialStr(dj)})
	}

	return menus
}

// GetViewFromDjCate 分类列表获取View
func GetViewFromDjCate(categories []ds.DjCategory) []model.MenuItem {
	var menus []model.MenuItem
	for _, category := range categories {
		menus = append(menus, model.MenuItem{Title: ReplaceSpecialStr(category.Name)})
	}

	return menus
}
