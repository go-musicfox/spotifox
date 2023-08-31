package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/storage"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/skratchdot/open-golang/open"
)

type LastfmAuth struct {
	baseMenu
	token string
	url   string
	err   error
}

func NewLastfmAuth(base baseMenu) *LastfmAuth {
	return &LastfmAuth{baseMenu: base}
}

func (m *LastfmAuth) GetMenuKey() string {
	return "last_fm_auth"
}

func (m *LastfmAuth) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "已点击，继续授权"},
	}
}

func (m *LastfmAuth) BeforeBackMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		m.token, m.url, m.err = "", "", nil
		return true, nil
	}
}

func (m *LastfmAuth) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		m.token, m.url, m.err = m.netease.lastfm.GetAuthUrlWithToken()
		if m.url != "" {
			_ = open.Start(m.url)
		}
		utils.Logger().Println("[lastfm] auth url: " + m.url)
		return true, nil
	}
}

func (m *LastfmAuth) SubMenu(mod_el *model.App, _ int) model.Menu {
	var err error

	loading := NewLoading(m.netease)
	loading.start()

	if m.netease.lastfmUser == nil {
		m.netease.lastfmUser = &storage.LastfmUser{}
	}
	m.netease.lastfmUser.SessionKey, err = m.netease.lastfm.GetSession(m.token)
	if err != nil {
		loading.complete()
		return NewLastfmRes(m.baseMenu, "授权", err, 1)
	}
	user, err := m.netease.lastfm.GetUserInfo(map[string]interface{}{})
	loading.complete()
	if err != nil {
		return NewLastfmRes(m.baseMenu, "授权", err, 1)
	}
	m.netease.lastfmUser.Id = user.Id
	m.netease.lastfmUser.Name = user.Name
	m.netease.lastfmUser.RealName = user.RealName
	m.netease.lastfmUser.Url = user.Url
	m.netease.lastfmUser.Store()
	return NewLastfmRes(m.baseMenu, "授权", nil, 3)
}

func (m *LastfmAuth) FormatMenuItem(item *model.MenuItem) {
	if m.err != nil {
		item.Subtitle = "[错误: " + m.err.Error() + "]"
		return
	}
	if m.url != "" {
		item.Subtitle = "打开链接进行授权"
		return
	}
	item.Subtitle = ""
}
