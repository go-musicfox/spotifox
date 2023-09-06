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
	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/go-musicfox/spotifox/internal/lastfm"
	"github.com/go-musicfox/spotifox/internal/player"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/internal/structs"
	"github.com/go-musicfox/spotifox/utils"
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
	// n.search = NewSearchPage(n)

	if configs.ConfigRegistry.ShowLyric && configs.ConfigRegistry.SpotifyCookie != "" {
		s.lyricClient = lyricsapi.NewLyricsApi(configs.ConfigRegistry.SpotifyCookie)
	}

	return s
}

// ToLoginPage
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

// ToSearchPage
func (s *Spotifox) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
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

		// 获取扩展信息
		{
			var (
				extInfo    storage.ExtInfo
				needUpdate = true
			)
			jsonStr, _ := table.GetByKVModel(extInfo)
			if len(jsonStr) != 0 {
				if err := json.Unmarshal(jsonStr, &extInfo); err == nil && semver.Compare(extInfo.StorageVersion, constants.AppVersion) >= 0 {
					needUpdate = false
				}
			}
			if needUpdate {
				localDir := utils.GetLocalDataDir()

				// refresh notifier
				_ = os.RemoveAll(path.Join(localDir, "musicfox-notifier.app"))

				// refresh logo
				_ = os.Remove(path.Join(localDir, constants.DefaultNotifyIcon))

				extInfo.StorageVersion = constants.AppVersion
				_ = table.SetByKVModel(extInfo, extInfo)
			}
		}

		// 检查更新
		if config.StartupCheckUpdate {
			if ok, newVersion := utils.CheckUpdate(); ok {
				if runtime.GOOS == "windows" {
					s.MustMain().EnterMenu(
						NewCheckUpdateMenu(newBaseMenu(s)),
						&model.MenuItem{Title: "新版本: " + newVersion, Subtitle: "当前版本: " + constants.AppVersion},
					)
				}

				utils.Notify(utils.NotifyContent{
					Title: "发现新版本: " + newVersion,
					Text:  "去看看呗",
					Url:   constants.AppLatestReleases,
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
