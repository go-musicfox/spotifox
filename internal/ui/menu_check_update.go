package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/locale"

	"github.com/skratchdot/open-golang/open"
)

type CheckUpdateMenu struct {
	baseMenu
	hasUpdate bool
}

func NewCheckUpdateMenu(base baseMenu) *CheckUpdateMenu {
	return &CheckUpdateMenu{
		baseMenu: base,
	}
}

func (m *CheckUpdateMenu) GetMenuKey() string {
	return "check_update"
}

func (m *CheckUpdateMenu) MenuViews() []model.MenuItem {
	if m.hasUpdate {
		return []model.MenuItem{
			{Title: locale.MustT("has_new_version"), Subtitle: "ENTER"},
		}
	}

	return []model.MenuItem{
		{Title: locale.MustT("has_no_new_version")},
	}
}

func (m *CheckUpdateMenu) SubMenu(_ *model.App, _ int) model.Menu {
	if m.hasUpdate {
		_ = open.Start(types.AppGithubUrl)
	}
	return nil
}

func (m *CheckUpdateMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		m.hasUpdate, _ = utils.CheckUpdate()
		return true, nil
	}
}
