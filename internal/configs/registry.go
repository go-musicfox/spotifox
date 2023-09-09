package configs

import (
	"runtime"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/spotifox/internal/types"

	"github.com/gookit/ini/v2"
)

var ConfigRegistry *Registry

type Registry struct {
	Startup  StartupOptions
	Progress ProgressOptions
	Spotify  SpotifyOptions
	Main     MainOptions
	Player   PlayerOptions
}

func (r *Registry) FillToModelOpts(opts *model.Options) {
	opts.StartupOptions = r.Startup.StartupOptions
	opts.ProgressOptions = r.Progress.ProgressOptions

	opts.AppName = types.AppName
	opts.WhetherDisplayTitle = r.Main.ShowTitle
	opts.PrimaryColor = r.Main.PrimaryColor
	opts.DualColumn = r.Main.DualColumn

	if r.Main.EnableMouseEvent {
		opts.TeaOptions = append(opts.TeaOptions, tea.WithMouseCellMotion())
	}
	if r.Main.AltScreen {
		opts.TeaOptions = append(opts.TeaOptions, tea.WithAltScreen())
	}
}

func NewRegistryWithDefault() *Registry {
	registry := &Registry{
		Startup: StartupOptions{
			StartupOptions: model.StartupOptions{
				EnableStartup:     true,
				LoadingDuration:   time.Second * types.StartupLoadingSeconds,
				TickDuration:      types.StartupTickDuration,
				ProgressOutBounce: true,
				Welcome:           types.AppName,
			},
			CheckUpdate: true,
		},
		Progress: ProgressOptions{
			ProgressOptions: model.ProgressOptions{
				FirstEmptyChar: []rune(types.ProgressEmptyChar)[0],
				EmptyChar:      []rune(types.ProgressEmptyChar)[0],
				LastEmptyChar:  []rune(types.ProgressEmptyChar)[0],
				FirstFullChar:  []rune(types.ProgressFullChar)[0],
				FullChar:       []rune(types.ProgressFullChar)[0],
				LastFullChar:   []rune(types.ProgressFullChar)[0],
			},
		},
		Main: MainOptions{
			Language:         "en",
			ShowTitle:        true,
			SongFormat:       Ogg320,
			PrimaryColor:     types.AppPrimaryColor,
			ShowLyric:        true,
			ShowLyricTrans:   true,
			ShowNotify:       true,
			NotifyIcon:       types.DefaultNotifyIcon,
			PProfPort:        types.MainPProfPort,
			AltScreen:        true,
			EnableMouseEvent: true,
		},
		Player: PlayerOptions{
			Engine: types.BeepPlayer,
		},
	}

	if runtime.GOOS == "darwin" {
		registry.Player.Engine = types.OsxPlayer
	}

	return registry
}

func NewRegistryFromIniFile(filepath string) *Registry {
	registry := NewRegistryWithDefault()

	if err := ini.LoadExists(filepath); err != nil {
		return registry
	}

	registry.Startup.EnableStartup = ini.Bool("startup.enable", true)
	registry.Startup.ProgressOutBounce = ini.Bool("startup.progressOutBounce", true)
	registry.Startup.LoadingDuration = time.Second * time.Duration(ini.Int("startup.loadingSeconds", types.StartupLoadingSeconds))
	registry.Startup.Welcome = ini.String("startup.welcome", types.AppName)
	registry.Startup.CheckUpdate = ini.Bool("startup.checkUpdate", true)

	emptyChar := ini.String("progress.emptyChar", types.ProgressEmptyChar)
	registry.Progress.EmptyChar = firstCharOrDefault(emptyChar, types.ProgressEmptyChar)
	firstEmptyChar := ini.String("progress.firstEmptyChar", types.ProgressEmptyChar)
	registry.Progress.FirstEmptyChar = firstCharOrDefault(firstEmptyChar, types.ProgressEmptyChar)
	lastEmptyChar := ini.String("progress.lastEmptyChar", types.ProgressEmptyChar)
	registry.Progress.LastEmptyChar = firstCharOrDefault(lastEmptyChar, types.ProgressEmptyChar)

	fullChar := ini.String("progress.fullChar", types.ProgressFullChar)
	registry.Progress.FullChar = firstCharOrDefault(fullChar, types.ProgressFullChar)
	firstFullChar := ini.String("progress.firstFullChar", types.ProgressFullChar)
	registry.Progress.FirstFullChar = firstCharOrDefault(firstFullChar, types.ProgressFullChar)
	lastFullChar := ini.String("progress.lastFullChar", types.ProgressFullChar)
	registry.Progress.LastFullChar = firstCharOrDefault(lastFullChar, types.ProgressFullChar)

	registry.Spotify.ClientId = types.SpotifyClientId
	if clientId := ini.Get("spotify.clientId"); clientId != "" {
		registry.Spotify.ClientId = clientId
	}
	registry.Spotify.Cookie = ini.Get("spotify.cookie", "")

	registry.Main.ShowTitle = ini.Bool("main.showTitle", true)
	songFormat := SongFormat(ini.String("main.songFormat", string(Ogg320)))
	if songFormat.IsValid() {
		registry.Main.SongFormat = songFormat
	}
	primaryColor := ini.String("main.primaryColor", types.AppPrimaryColor)
	if primaryColor != "" {
		registry.Main.PrimaryColor = primaryColor
	} else {
		registry.Main.PrimaryColor = types.AppPrimaryColor
	}
	registry.Main.Language = ini.String("main.language", "en")
	registry.Main.NotifyIcon = ini.String("main.notifyIcon", types.DefaultNotifyIcon)
	registry.Main.ShowLyric = ini.Bool("main.showLyric", true)
	registry.Main.LyricOffset = ini.Int("main.lyricOffset", 0)
	registry.Main.ShowLyricTrans = ini.Bool("main.showLyricTrans", false)
	registry.Main.ShowNotify = ini.Bool("main.enableNotify", true)
	registry.Main.PProfPort = ini.Int("main.pprofPort", types.MainPProfPort)
	registry.Main.AltScreen = ini.Bool("main.altScreen", true)
	registry.Main.EnableMouseEvent = ini.Bool("main.enableMouseEvent", true)
	registry.Main.DualColumn = ini.Bool("main.dualColumn", true)

	registry.Main.LastfmKey = types.LastfmKey
	if key := ini.String("main.lastfmKey"); key != "" {
		registry.Main.LastfmKey = key
	}
	registry.Main.LastfmSecret = types.LastfmSecret
	if secret := ini.String("main.lastfmSecret"); secret != "" {
		registry.Main.LastfmSecret = secret
	}

	defaultPlayer := types.BeepPlayer
	if runtime.GOOS == "darwin" {
		defaultPlayer = types.OsxPlayer
	}
	registry.Player.Engine = ini.String("player.engine", defaultPlayer)

	return registry
}

func firstCharOrDefault(s, defaultStr string) rune {
	if len(s) > 0 {
		return []rune(s)[0]
	}
	return []rune(defaultStr)[0]
}
