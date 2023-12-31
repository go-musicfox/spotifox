package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/locale"

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
		{Title: locale.MustT("already_click")},
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
		m.token, m.url, m.err = m.spotifox.lastfm.GetAuthUrlWithToken()
		if m.url != "" {
			_ = open.Start(m.url)
		}
		utils.Logger().Println("[lastfm] auth url: " + m.url)
		return true, nil
	}
}

func (m *LastfmAuth) SubMenu(mod_el *model.App, _ int) model.Menu {
	var err error

	loading := model.NewLoading(m.spotifox.MustMain())
	loading.Start()

	if m.spotifox.lastfmUser == nil {
		m.spotifox.lastfmUser = &storage.LastfmUser{}
	}
	m.spotifox.lastfmUser.SessionKey, err = m.spotifox.lastfm.GetSession(m.token)
	if err != nil {
		loading.Complete()
		return NewLastfmRes(m.baseMenu, locale.MustT("auth"), err, 1)
	}
	user, err := m.spotifox.lastfm.GetUserInfo(map[string]interface{}{})
	loading.Complete()
	if err != nil {
		return NewLastfmRes(m.baseMenu, locale.MustT("auth"), err, 1)
	}
	m.spotifox.lastfmUser.Id = user.Id
	m.spotifox.lastfmUser.Name = user.Name
	m.spotifox.lastfmUser.RealName = user.RealName
	m.spotifox.lastfmUser.Url = user.Url
	m.spotifox.lastfmUser.Store()
	return NewLastfmRes(m.baseMenu, locale.MustT("auth"), nil, 3)
}

func (m *LastfmAuth) FormatMenuItem(item *model.MenuItem) {
	if m.err != nil {
		item.Subtitle = "[" + locale.MustT("error") + ": " + m.err.Error() + "]"
		return
	}
	if m.url != "" {
		item.Subtitle = locale.MustT("open_url_to_auth")
		return
	}
	item.Subtitle = ""
}
