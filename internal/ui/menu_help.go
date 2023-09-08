package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils/locale"

	"github.com/skratchdot/open-golang/open"
)

type HelpMenu struct {
	baseMenu
	menus []model.MenuItem
}

func NewHelpMenu(base baseMenu) *HelpMenu {
	menu := &HelpMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: locale.MustT("star_me")},
			{Title: "SPACE", Subtitle: locale.MustT("play_paused")},
			{Title: "h/H/LEFT", Subtitle: locale.MustT("move_left")},
			{Title: "l/L/RIGHT", Subtitle: locale.MustT("move_right")},
			{Title: "k/K/UP", Subtitle: locale.MustT("move_up")},
			{Title: "j/J/DOWN", Subtitle: locale.MustT("move_down")},
			{Title: "g", Subtitle: locale.MustT("move_top")},
			{Title: "G", Subtitle: locale.MustT("move_bottom")},
			{Title: "[", Subtitle: locale.MustT("pre_track")},
			{Title: "]", Subtitle: locale.MustT("next_track")},
			{Title: "-", Subtitle: locale.MustT("up_volume")},
			{Title: "=", Subtitle: locale.MustT("down_volume")},
			{Title: "n/N/ENTER", Subtitle: locale.MustT("enter")},
			{Title: "b/B/ESC", Subtitle: locale.MustT("back")},
			{Title: "q/Q", Subtitle: locale.MustT("quit")},
			{Title: "w/W", Subtitle: locale.MustT("logout_and_quit")},
			{Title: "p/P", Subtitle: locale.MustT("switch_play_mode")},
			{Title: ",", Subtitle: locale.MustT("like_playing_track")},
			{Title: "<", Subtitle: locale.MustT("like_selected_track")},
			{Title: ".", Subtitle: locale.MustT("dislike_playing_track")},
			{Title: ">", Subtitle: locale.MustT("dislike_selected_track")},
			{Title: "`", Subtitle: locale.MustT("add_playing_track_to_playlist")},
			{Title: "Tab", Subtitle: locale.MustT("add_selected_track_to_playlist")},
			{Title: "~", Subtitle: locale.MustT("remove_playing_track_from_playlist")},
			{Title: "Shift+Tab", Subtitle: locale.MustT("remove_selected_track_from_playlist")},
			{Title: "d", Subtitle: locale.MustT("download_playing_track")},
			{Title: "D", Subtitle: locale.MustT("download_selected_track")},
			{Title: "c/C", Subtitle: locale.MustT("current_playlist")},
			{Title: "r/R", Subtitle: locale.MustT("rerender_ui")},
			{Title: "/", Subtitle: locale.MustT("search_cur_menulist")},
			{Title: "?", Subtitle: locale.MustT("help")},
			{Title: "a", Subtitle: locale.MustT("album_of_playing_track")},
			{Title: "A", Subtitle: locale.MustT("album_of_selected_track")},
			{Title: "s", Subtitle: locale.MustT("artist_of_playing_track")},
			{Title: "S", Subtitle: locale.MustT("artist_of_selected_track")},
			{Title: "o", Subtitle: locale.MustT("open_playing_track_url")},
			{Title: "O", Subtitle: locale.MustT("open_selected_item_url")},
			{Title: ";/:", Subtitle: locale.MustT("follow_selected_playlist")},
			{Title: "'/\"", Subtitle: locale.MustT("unfollow_selected_playlist")},
		},
	}

	return menu
}

func (m *HelpMenu) GetMenuKey() string {
	return "help_menu"
}

func (m *HelpMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *HelpMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index == 0 {
		_ = open.Start(types.AppGithubUrl)
	}
	return nil
}
