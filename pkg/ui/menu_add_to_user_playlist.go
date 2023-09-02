package ui

import (
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/structs"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type AddToUserPlaylistMenu struct {
	baseMenu
	menus     []model.MenuItem
	playlists []spotify.SimplePlaylist
	song      structs.Song
	offset    int
	limit     int
	hasMore   bool
	action    bool // true for add, false for del
}

func NewAddToUserPlaylistMenu(base baseMenu, song structs.Song, action bool) *AddToUserPlaylistMenu {
	return &AddToUserPlaylistMenu{
		baseMenu: base,
		offset:   0,
		limit:    100,
		action:   action,
		song:     song,
	}
}

func (m *AddToUserPlaylistMenu) IsSearchable() bool {
	return true
}

func (m *AddToUserPlaylistMenu) GetMenuKey() string {
	return "add_to_user_playlist_" + m.spotifox.user.ID
}

func (m *AddToUserPlaylistMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *AddToUserPlaylistMenu) Playlists() []spotify.SimplePlaylist {
	return m.playlists
}

func (m *AddToUserPlaylistMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *AddToUserPlaylistMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 等于0，获取当前用户歌单
		if utils.CheckUserInfo(m.spotifox.user) == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		userId := m.spotifox.user.ID

		userPlaylists := service.UserPlaylistService{
			Uid:    userId,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		var menus []model.MenuItem
		// m.playlists = utils.GetPlaylists(response)
		for _, playlist := range m.playlists {
			menus = append(menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		}
		m.menus = menus

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true, nil
	}
}

func (m *AddToUserPlaylistMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		userId := m.spotifox.user.ID

		m.offset = m.offset + len(m.menus)
		userPlaylists := service.UserPlaylistService{
			Uid:    userId,
			Limit:  strconv.Itoa(m.limit),
			Offset: strconv.Itoa(m.offset),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			page, _ := m.spotifox.ToLoginPage(nil)
			return false, page
		} else if codeType != utils.Success {
			return false, nil
		}

		// list := utils.GetPlaylists(response)
		// for _, playlist := range list {
		// 	m.menus = append(m.menus, model.MenuItem{Title: utils.ReplaceSpecialStr(playlist.Name)})
		// }

		// m.playlists = append(m.playlists, list...)

		// 是否有更多
		if hasMore, err := jsonparser.GetBoolean(response, "more"); err == nil {
			m.hasMore = hasMore
		}

		return true, nil
	}
}
