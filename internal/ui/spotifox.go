package ui

import (
	"encoding/json"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/anhoder/foxful-cli/model"
	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/lastfm"
	"github.com/go-musicfox/spotifox/internal/player"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/internal/structs"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/locale"
	lyricsapi "github.com/raitonoberu/lyricsapi/lyrics"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/mod/semver"
)

type Spotifox struct {
	user       *structs.User
	lastfm     *lastfm.Client
	lastfmUser *storage.LastfmUser

	sess          respot.Session
	spotifyClient *spotify.Client
	lyricClient   *lyricsapi.LyricsApi

	*model.App
	login  *LoginPage
	search *SearchPage

	player *Player
}

func NewSpotifox(app *model.App) *Spotifox {
	var s = &Spotifox{
		lastfm: lastfm.NewClient(),
		sess:   NewSpotifySession(),
		App:    app,
	}
	s.player = NewPlayer(s)
	s.login = NewLoginPage(s)
	s.search = NewSearchPage(s)

	if configs.ConfigRegistry.Main.ShowLyric && configs.ConfigRegistry.Spotify.Cookie != "" {
		s.lyricClient = lyricsapi.NewLyricsApi(configs.ConfigRegistry.Spotify.Cookie)
	}

	return s
}

func (s *Spotifox) HandleResCode(code utils.ResCode, callback LoginCallback) (bool, model.Page) {
	utils.Logger().Printf("[INFO] code: %+v", code)
	switch code {
	case utils.NeedLogin:
		page, _ := s.ToLoginPage(callback)
		return true, page
	case utils.TokenExpired:
		// refresh token
		s.login.AfterLogin = callback
		page, _ := s.login.handleLoginSuccess()
		return true, page
	}
	return false, nil
}

func (s *Spotifox) ToLoginPage(callback LoginCallback) (model.Page, tea.Cmd) {
	s.login.AfterLogin = callback
	if s.user != nil && s.user.Username != "" && len(s.user.AuthBlob) > 0 {
		login := &s.sess.Context().Login
		login.Username = s.user.Username
		login.AuthData = s.user.AuthBlob
		err := s.ReconnSessionWhenNeed(func() error {
			return s.sess.Login()
		})
		if err == nil {
			return s.login.handleLoginSuccess()
		}
		utils.Logger().Printf("login by auth blob failed, err: %+v", err)
	}

	return s.login, tickLogin(time.Nanosecond)
}

func (s *Spotifox) ToSearchPage(searchType spotify.SearchType) (model.Page, tea.Cmd) {
	s.search.searchType = searchType
	return s.search, tickSearch(time.Nanosecond)
}

func (s *Spotifox) InitHook(_ *model.App) {
	config := configs.ConfigRegistry
	// projectDir := utils.GetLocalDataDir()

	// cookie jar
	// cookieJar, _ := cookiejar.NewFileJar(path.Join(projectDir, "cookie"), nil)
	// util.SetGlobalCookieJar(cookieJar)

	// DBManager init
	storage.DBManager = new(storage.LocalDBManager)

	go utils.PanicRecoverWrapper(false, func() {
		table := storage.NewTable()

		// get user info
		if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
			if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
				s.user = &user
			}
		}
		// refresh username
		s.MustMain().RefreshMenuTitle()

		// get user info of lastfm
		var lastfmUser storage.LastfmUser
		if jsonStr, err := table.GetByKVModel(&lastfmUser); err == nil {
			if err = json.Unmarshal(jsonStr, &lastfmUser); err == nil {
				s.lastfmUser = &lastfmUser
				s.lastfm.SetSession(lastfmUser.SessionKey)
			}
		}
		s.MustMain().RefreshMenuList()

		// get play mode
		if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
			var playMode player.Mode
			if err = json.Unmarshal(jsonStr, &playMode); err == nil {
				s.player.mode = playMode
			}
		}

		// get player volume
		if jsonStr, err := table.GetByKVModel(storage.Volume{}); err == nil && len(jsonStr) > 0 {
			var volume int
			if err = json.Unmarshal(jsonStr, &volume); err == nil {
				v, ok := s.player.Player.(storage.VolumeStorable)
				if ok {
					v.SetVolume(volume)
				}
			}
		}

		// get playing info
		if jsonStr, err := table.GetByKVModel(storage.PlayerSnapshot{}); err == nil && len(jsonStr) > 0 {
			var snapshot storage.PlayerSnapshot
			if err = json.Unmarshal(jsonStr, &snapshot); err == nil {
				p := s.player
				p.curSongIndex = snapshot.CurSongIndex
				p.playlist = snapshot.Playlist
				p.playlistUpdateAt = snapshot.PlaylistUpdateAt
				p.curSong = p.playlist[p.curSongIndex]
				p.isCurSongLiked = snapshot.IsCurSongLiked
				p.playingMenuKey = "from_local_db" // reset menu key
			}
		}
		s.Rerender(false)

		// get ext info
		{
			var (
				extInfo    storage.ExtInfo
				needUpdate = true
			)
			jsonStr, _ := table.GetByKVModel(extInfo)
			if len(jsonStr) != 0 {
				if err := json.Unmarshal(jsonStr, &extInfo); err == nil && semver.Compare(extInfo.StorageVersion, types.AppVersion) >= 0 {
					needUpdate = false
				}
			}
			if needUpdate {
				localDir := utils.GetLocalDataDir()

				// refresh notifier
				_ = os.RemoveAll(path.Join(localDir, "musicfox-notifier.app"))

				// refresh logo
				_ = os.Remove(path.Join(localDir, types.DefaultNotifyIcon))

				extInfo.StorageVersion = types.AppVersion
				_ = table.SetByKVModel(extInfo, extInfo)
			}
		}

		// check update
		if config.Startup.CheckUpdate {
			if ok, newVersion := utils.CheckUpdate(); ok {
				if runtime.GOOS == "windows" {
					s.MustMain().EnterMenu(
						NewCheckUpdateMenu(newBaseMenu(s)),
						&model.MenuItem{
							Title:    locale.MustT("new_version_title", locale.WithTplData(map[string]string{"NewVersion": newVersion})),
							Subtitle: locale.MustT("new_version_subtitle", locale.WithTplData(map[string]string{"CurVersion": types.AppVersion})),
						},
					)
				}

				utils.Notify(utils.NotifyContent{
					Title: locale.MustT("new_version_notify_title", locale.WithTplData(map[string]string{"NewVersion": newVersion})),
					Text:  locale.MustT("new_version_notify_txt"),
					Url:   types.AppLatestReleases,
				})
			}
		}
	})
}

func (s *Spotifox) CloseHook(_ *model.App) {
	s.player.Close()
}

func (s *Spotifox) Player() *Player {
	return s.player
}
