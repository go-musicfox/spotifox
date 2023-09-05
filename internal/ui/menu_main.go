package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

type MainMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
}

func NewMainMenu(netease *Spotifox) *MainMenu {
	base := newBaseMenu(netease)
	mainMenu := &MainMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "我喜欢的音乐"},
			{Title: "关注的歌单"},
			{Title: "帮助"},
			{Title: "检查更新"},
		},
		menuList: []Menu{
			NewLikedSongsMenu(base),
			NewUserPlaylistMenu(base, CurUser),
			NewHelpMenu(base),
			NewCheckUpdateMenu(base),
		},
	}
	return mainMenu
}

func (m *MainMenu) FormatMenuItem(item *model.MenuItem) {
	if m.spotifox.user == nil {
		item.Subtitle = "[未登录]"
		return
	}
	if m.spotifox.user.DisplayName != "" {
		item.Subtitle = "[" + m.spotifox.user.DisplayName + "]"
		return
	}
	item.Subtitle = "[" + m.spotifox.user.Username + "]"
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	for i, menu := range m.menuList {
		menu.FormatMenuItem(&m.menus[i])
	}
	return m.menus
}

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
