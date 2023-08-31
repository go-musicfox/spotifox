package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/structs"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/go-musicfox/netease-music/service"
)

type PersonalFmMenu struct {
	baseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewPersonalFmMenu(base baseMenu) *PersonalFmMenu {
	return &PersonalFmMenu{
		baseMenu: base,
	}
}

func (m *PersonalFmMenu) IsSearchable() bool {
	return true
}

func (m *PersonalFmMenu) IsPlayable() bool {
	return true
}

func (m *PersonalFmMenu) GetMenuKey() string {
	return "personal_fm"
}

func (m *PersonalFmMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *PersonalFmMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		// 已有数据
		if len(m.menus) > 0 && len(m.songs) > 0 {
			return true, nil
		}

		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}

		// 响应中获取数据
		m.songs = utils.GetFmSongs(response)
		m.menus = utils.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *PersonalFmMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		personalFm := service.PersonalFmService{}
		code, response := personalFm.PersonalFm()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false, nil
		}
		songs := utils.GetFmSongs(response)
		menus := utils.GetViewFromSongs(songs)

		m.menus = append(m.menus, menus...)
		m.songs = append(m.songs, songs...)
		m.netease.player.playlist = m.songs
		m.netease.player.playlistUpdateAt = time.Now()

		return true, nil
	}
}

func (m *PersonalFmMenu) Songs() []structs.Song {
	return m.songs
}
