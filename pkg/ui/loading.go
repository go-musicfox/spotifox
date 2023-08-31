package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/configs"

	"github.com/muesli/termenv"
)

type Loading struct {
	netease   *Spotifox
	menuTitle *model.MenuItem
}

func NewLoading(n *Spotifox, menuTitle ...*model.MenuItem) *Loading {
	l := &Loading{netease: n}
	if len(menuTitle) > 0 {
		l.menuTitle = menuTitle[0]
	}
	return l
}

// 开始
func (loading *Loading) start() {
	var (
		main      = loading.netease.App.MustMain()
		subTitle  string
		menuTitle *model.MenuItem
	)
	termenv.DefaultOutput().MoveCursor(main.MenuTitleStartRow(), 0)

	if loading.menuTitle != nil {
		menuTitle = loading.menuTitle
	} else {
		menuTitle = main.MenuTitle()
	}

	if menuTitle.Subtitle != "" {
		subTitle = menuTitle.Subtitle + " " + configs.ConfigRegistry.LoadingText
	} else {
		subTitle = configs.ConfigRegistry.LoadingText
	}
	fmt.Print(main.MenuTitleView(loading.netease.App, nil, &model.MenuItem{
		Title:    menuTitle.Title,
		Subtitle: subTitle,
	}))

	termenv.DefaultOutput().MoveCursor(0, 0)
}

// 完成
func (loading *Loading) complete() {
	var (
		main      = loading.netease.App.MustMain()
		menuTitle *model.MenuItem
	)
	termenv.DefaultOutput().MoveCursor(main.MenuTitleStartRow(), 0)

	if loading.menuTitle != nil {
		menuTitle = loading.menuTitle
	} else {
		menuTitle = main.MenuTitle()
	}
	fmt.Print(main.MenuTitleView(loading.netease.App, nil, menuTitle))

	termenv.DefaultOutput().MoveCursor(0, 0)
}
