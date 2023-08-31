package ui

import (
	"encoding/json"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/spotifox/pkg/configs"
	"github.com/go-musicfox/spotifox/pkg/constants"
	"github.com/go-musicfox/spotifox/pkg/lastfm"
	"github.com/go-musicfox/spotifox/pkg/player"
	"github.com/go-musicfox/spotifox/pkg/storage"
	"github.com/go-musicfox/spotifox/pkg/structs"
	"github.com/go-musicfox/spotifox/utils"
	"golang.org/x/mod/semver"
)

type Spotifox struct {
	user       *structs.User
	lastfm     *lastfm.Client
	lastfmUser *storage.LastfmUser

	*model.App
	login  *LoginPage
	search *SearchPage

	player *Player
}

func NewSpotifox(app *model.App) *Spotifox {
	n := new(Spotifox)
	n.lastfm = lastfm.NewClient()
	n.player = NewPlayer(n)
	n.login = NewLoginPage(n)
	// n.search = NewSearchPage(n)
	n.App = app

	return n
}

// ToLoginPage
func (n *Spotifox) ToLoginPage(callback func(newMenu model.Menu, newTitle *model.MenuItem) model.Page) (model.Page, tea.Cmd) {
	//n.login.AfterLogin = callback
	return n.login, tickLogin(time.Nanosecond)
}

// ToSearchPage
func (n *Spotifox) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
	n.search.searchType = searchType
	return n.search, tickSearch(time.Nanosecond)
}

func (n *Spotifox) InitHook(_ *model.App) {
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
				n.user = &user
			}
		}
		// refresh username
		n.MustMain().RefreshMenuTitle()

		// get user info of lastfm
		var lastfmUser storage.LastfmUser
		if jsonStr, err := table.GetByKVModel(&lastfmUser); err == nil {
			if err = json.Unmarshal(jsonStr, &lastfmUser); err == nil {
				n.lastfmUser = &lastfmUser
				n.lastfm.SetSession(lastfmUser.SessionKey)
			}
		}
		n.MustMain().RefreshMenuList()

		// get play mode
		if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
			var playMode player.Mode
			if err = json.Unmarshal(jsonStr, &playMode); err == nil {
				n.player.mode = playMode
			}
		}

		// get player volume
		if jsonStr, err := table.GetByKVModel(storage.Volume{}); err == nil && len(jsonStr) > 0 {
			var volume int
			if err = json.Unmarshal(jsonStr, &volume); err == nil {
				v, ok := n.player.Player.(storage.VolumeStorable)
				if ok {
					v.SetVolume(volume)
				}
			}
		}

		// get playing info
		if jsonStr, err := table.GetByKVModel(storage.PlayerSnapshot{}); err == nil && len(jsonStr) > 0 {
			var snapshot storage.PlayerSnapshot
			if err = json.Unmarshal(jsonStr, &snapshot); err == nil {
				p := n.player
				p.curSongIndex = snapshot.CurSongIndex
				p.playlist = snapshot.Playlist
				p.playlistUpdateAt = snapshot.PlaylistUpdateAt
				p.curSong = p.playlist[p.curSongIndex]
				p.playingMenuKey = "from_local_db" // reset menu key
			}
		}
		n.Rerender(false)

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
					n.MustMain().EnterMenu(
						NewCheckUpdateMenu(newBaseMenu(n)),
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

func (n *Spotifox) CloseHook(_ *model.App) {
	n.player.Close()
}

func (n *Spotifox) Player() *Player {
	return n.player
}
