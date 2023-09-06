package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"

	"github.com/go-musicfox/netease-music/service"
)

type SearchResultMenu struct {
	baseMenu
	menus      []model.MenuItem
	offset     int
	searchType SearchType
	keyword    string
	result     interface{}
}

var playableTypes = map[SearchType]bool{
	StSingleSong: true,
	StAlbum:      false,
	StSinger:     false,
	StPlaylist:   false,
	StUser:       false,
	StLyric:      true,
	StRadio:      false,
}

func NewSearchResultMenu(base baseMenu, searchType SearchType) *SearchResultMenu {
	return &SearchResultMenu{
		baseMenu:   base,
		offset:     0,
		searchType: searchType,
	}
}

func (m *SearchResultMenu) IsSearchable() bool {
	return true
}

func (m *SearchResultMenu) BeforeBackMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.search.wordsInput.Value() != "" {
			m.spotifox.search.wordsInput.SetValue("")
		}

		return true, nil
	}
}

func (m *SearchResultMenu) IsPlayable() bool {
	return playableTypes[m.searchType]
}

func (m *SearchResultMenu) GetMenuKey() string {
	return fmt.Sprintf("search_result_%d_%s", m.searchType, m.keyword)
}

func (m *SearchResultMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SearchResultMenu) SubMenu(_ *model.App, index int) model.Menu {
	switch resultWithType := m.result.(type) {
	case []spotify.FullTrack:
		return nil
	case []spotify.SimpleAlbum:
		if index >= len(resultWithType) {
			return nil
		}
		return NewAlbumDetailMenu(m.baseMenu, resultWithType[index])
	case []spotify.SimplePlaylist:
		if index >= len(resultWithType) {
			return nil
		}
		return NewPlaylistDetailMenu(m.baseMenu, resultWithType[index].ID)
	case []spotify.SimpleArtist:
		if index >= len(resultWithType) {
			return nil
		}
		return NewArtistDetailMenu(m.baseMenu, resultWithType[index].ID, resultWithType[index].Name)
	case []spotify.User:
		if index >= len(resultWithType) {
			return nil
		}
		return NewUserPlaylistMenu(m.baseMenu, resultWithType[index].ID)
	}

	return nil
}

func (m *SearchResultMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		var (
			code     float64
			response []byte
		)
		m.offset += constants.SearchPageSize
		searchService := service.SearchService{
			S:      m.keyword,
			Type:   strconv.Itoa(int(m.searchType)),
			Limit:  strconv.Itoa(constants.SearchPageSize),
			Offset: strconv.Itoa(m.offset),
		}
		code, response = searchService.Search()

		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			m.offset -= constants.SearchPageSize
			return false, nil
		}

		m.appendResult(response)
		m.convertMenus()
		return true, nil
	}
}

func (m *SearchResultMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.spotifox.search.wordsInput.Value() == "" {
			// 显示搜索页面
			page, _ := m.spotifox.ToSearchPage(m.searchType)
			return false, page
		}

		m.result = m.spotifox.search.result
		m.searchType = m.spotifox.search.searchType
		m.keyword = m.spotifox.search.wordsInput.Value()
		m.convertMenus()
		return true, nil
	}
}

func (m *SearchResultMenu) appendResult(response []byte) {
	switch m.searchType {
	case StSingleSong, StLyric:
		songs, _ := m.result.([]spotify.FullTrack)
		// songs = append(songs, appendSongs...)
		m.result = songs
	case StAlbum:
		albums, _ := m.result.([]spotify.SimpleAlbum)
		// albums = append(albums, appendAlbums...)
		m.result = albums
	case StSinger:
		artists, _ := m.result.([]spotify.SimpleArtist)
		// artists = append(artists, appendArtists...)
		m.result = artists
	case StPlaylist:
		playlists, _ := m.result.([]spotify.SimplePlaylist)
		// playlists = append(playlists, appendPlaylists...)
		m.result = playlists
	case StUser:
		users, _ := m.result.([]spotify.User)
		// users = append(users, appendUsers...)
		m.result = users
	}
}

func (m *SearchResultMenu) convertMenus() {
	// switch resultWithType := m.result.(type) {
	// case []spotify.FullTrack:
	// 	m.menus = utils.GetViewFromSongs(resultWithType)
	// case []spotify.SimpleAlbum:
	// 	m.menus = utils.GetViewFromAlbums(resultWithType)
	// case []spotify.SimplePlaylist:
	// 	m.menus = utils.GetViewFromPlaylists(resultWithType)
	// case []spotify.SimpleArtist:
	// 	m.menus = utils.GetViewFromArtists(resultWithType)
	// case []spotify.User:
	// 	m.menus = utils.GetViewFromUsers(resultWithType)
	// }
}

func (m *SearchResultMenu) Songs() []spotify.FullTrack {
	if songs, ok := m.result.([]spotify.FullTrack); ok {
		return songs
	}
	return nil
}

func (m *SearchResultMenu) Playlists() []spotify.SimplePlaylist {
	if playlists, ok := m.result.([]spotify.SimplePlaylist); ok {
		return playlists
	}
	return nil
}

func (m *SearchResultMenu) Albums() []spotify.SimpleAlbum {
	if albums, ok := m.result.([]spotify.SimpleAlbum); ok {
		return albums
	}
	return nil
}

func (m *SearchResultMenu) Artists() []spotify.SimpleArtist {
	if artists, ok := m.result.([]spotify.SimpleArtist); ok {
		return artists
	}
	return nil
}
