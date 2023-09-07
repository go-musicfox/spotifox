package ui

import (
	"context"
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type SearchResultMenu struct {
	baseMenu
	menus      []model.MenuItem
	offset     int
	searchType spotify.SearchType
	keyword    string
	result     any
}

var playableTypes = map[spotify.SearchType]bool{
	spotify.SearchTypeTrack:    true,
	spotify.SearchTypeAlbum:    false,
	spotify.SearchTypeArtist:   false,
	spotify.SearchTypePlaylist: false,
	// spotify.SearchTypeShow:     false,
	// spotify.SearchTypeEpisode:  false,
}

func NewSearchResultMenu(base baseMenu, searchType spotify.SearchType) *SearchResultMenu {
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
	}

	return nil
}

func (m *SearchResultMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		m.offset += constants.SearchPageSize

		res, err := m.spotifox.spotifyClient.Search(context.Background(), m.keyword, m.searchType, spotify.Limit(constants.SearchPageSize), spotify.Offset(m.offset))
		if utils.CheckSpotifyErr(err) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(BottomOutHookCallback(main, m))
			return false, page
		}
		if err != nil {
			return m.handleFetchErr(errors.Wrap(err, "search item failed"))
		}

		switch m.searchType {
		case spotify.SearchTypeTrack:
			var tracks, _ = m.result.([]spotify.FullTrack)
			m.result = append(tracks, res.Tracks.Tracks...)
		case spotify.SearchTypeAlbum:
			var albums, _ = m.result.([]spotify.SimpleAlbum)
			m.result = append(albums, res.Albums.Albums...)
		case spotify.SearchTypeArtist:
			var artists, _ = m.result.([]spotify.SimpleArtist)
			for _, artist := range res.Artists.Artists {
				artists = append(artists, artist.SimpleArtist)
			}
			m.result = artists
		case spotify.SearchTypePlaylist:
			var playlists, _ = m.result.([]spotify.SimplePlaylist)
			m.result = append(playlists, res.Playlists.Playlists...)
		case spotify.SearchTypeShow:
			var shows, _ = m.result.([]spotify.FullShow)
			m.result = append(shows, res.Shows.Shows...)
		case spotify.SearchTypeEpisode:
			var episodes, _ = m.result.([]spotify.EpisodePage)
			m.result = append(episodes, res.Episodes.Episodes...)
		}

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

func (m *SearchResultMenu) convertMenus() {
	switch resultWithType := m.result.(type) {
	case []spotify.FullTrack:
		m.menus = utils.MenuItemsFromSongs(resultWithType)
	case []spotify.SimpleAlbum:
		m.menus = utils.MenuItemsFromAlbums(resultWithType)
	case []spotify.SimplePlaylist:
		m.menus = utils.MenuItemsFromPlaylists(resultWithType)
	case []spotify.SimpleArtist:
		m.menus = utils.MenuItemsFromArtists(resultWithType)
	}
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
