package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/utils/locale"

	"github.com/skratchdot/open-golang/open"
)

type Lastfm struct {
	baseMenu
	auth *LastfmAuth
}

func NewLastfm(base baseMenu) *Lastfm {
	return &Lastfm{
		baseMenu: base,
		auth:     NewLastfmAuth(base),
	}
}

func (m *Lastfm) GetMenuKey() string {
	return "last_fm"
}

func (m *Lastfm) MenuViews() []model.MenuItem {
	if m.spotifox.lastfmUser == nil || m.spotifox.lastfmUser.SessionKey == "" {
		return []model.MenuItem{
			{Title: locale.MustT("to_auth")},
		}
	}
	return []model.MenuItem{
		{Title: locale.MustT("view_user_info")},
		{Title: locale.MustT("clear_auth")},
	}
}

func (m *Lastfm) SubMenu(_ *model.App, index int) model.Menu {
	if m.spotifox.lastfmUser == nil || m.spotifox.lastfmUser.SessionKey == "" {
		return m.auth
	}
	switch index {
	case 0:
		_ = open.Start(m.spotifox.lastfmUser.Url)
	case 1:
		m.spotifox.lastfmUser = &storage.LastfmUser{}
		m.spotifox.lastfmUser.Clear()
		return NewLastfmRes(m.baseMenu, locale.MustT("clear_auth"), nil, 2)
	}
	return nil
}

func (m *Lastfm) FormatMenuItem(item *model.MenuItem) {
	if m.spotifox.lastfmUser == nil || m.spotifox.lastfmUser.SessionKey == "" {
		item.Subtitle = "[" + locale.MustT("unauth") + "]"
	} else {
		item.Subtitle = fmt.Sprintf("[%s]", m.spotifox.lastfmUser.Name)
	}
}
