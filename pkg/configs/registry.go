package configs

import (
	"runtime"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/spotifox/pkg/constants"

	"github.com/go-musicfox/netease-music/service"
	"github.com/gookit/ini/v2"
)

var ConfigRegistry *Registry

type Registry struct {
	StartupShow              bool          // 显示启动页
	StartupProgressOutBounce bool          // 是否启动页进度条回弹效果
	StartupLoadingDuration   time.Duration // 启动页加载时长
	StartupWelcome           string        // 启动页欢迎语
	StartupCheckUpdate       bool          // 启动检查更新

	ProgressFirstEmptyChar rune // 进度条第一个未加载字符
	ProgressEmptyChar      rune // 进度条未加载字符
	ProgressLastEmptyChar  rune // 进度条最后一个未加载字符
	ProgressFirstFullChar  rune // 进度条第一个已加载字符
	ProgressFullChar       rune // 进度条已加载字符
	ProgressLastFullChar   rune // 进度条最后一个已加载字符

	SpotifyClientId  string
	ShowTitle        bool                     // 主界面是否显示标题
	LoadingText      string                   // 主页面加载中提示
	PlayerSongLevel  service.SongQualityLevel // 歌曲音质级别
	PrimaryColor     string                   // 主题色
	ShowLyric        bool                     // 显示歌词
	LyricOffset      int                      // 偏移:ms
	ShowLyricTrans   bool                     // 显示歌词翻译
	ShowNotify       bool                     // 显示通知
	NotifyIcon       string                   // logo 图片名
	PProfPort        int                      // pprof端口
	AltScreen        bool                     // AltScreen显示模式
	EnableMouseEvent bool                     // 启用鼠标事件
	DoubleColumn     bool                     // 是否双列显示

	PlayerEngine         string // 播放引擎
	PlayerBeepMp3Decoder string // beep mp3解码器
}

func (r *Registry) FillToModelOpts(opts *model.Options) {
	opts.StartupOptions.EnableStartup = r.StartupShow
	opts.StartupOptions.LoadingDuration = r.StartupLoadingDuration
	opts.StartupOptions.TickDuration = constants.StartupTickDuration
	opts.StartupOptions.ProgressOutBounce = r.StartupProgressOutBounce
	opts.StartupOptions.Welcome = r.StartupWelcome

	opts.ProgressOptions.FirstFullChar = r.ProgressFirstFullChar
	opts.ProgressOptions.FullChar = r.ProgressFullChar
	opts.ProgressOptions.LastFullChar = r.ProgressLastFullChar
	opts.ProgressOptions.FirstEmptyChar = r.ProgressFirstEmptyChar
	opts.ProgressOptions.EmptyChar = r.ProgressEmptyChar
	opts.ProgressOptions.LastEmptyChar = r.ProgressLastEmptyChar

	opts.AppName = constants.AppName
	opts.WhetherDisplayTitle = r.ShowTitle
	opts.LoadingText = r.LoadingText
	opts.PrimaryColor = r.PrimaryColor
	opts.DualColumn = r.DoubleColumn

	if r.EnableMouseEvent {
		opts.TeaOptions = append(opts.TeaOptions, tea.WithMouseCellMotion())
	}
	if r.AltScreen {
		opts.TeaOptions = append(opts.TeaOptions, tea.WithAltScreen())
	}
}

func NewRegistryWithDefault() *Registry {
	registry := &Registry{
		StartupShow:              true,
		StartupProgressOutBounce: true,
		StartupLoadingDuration:   time.Second * constants.StartupLoadingSeconds,
		StartupWelcome:           constants.AppName,
		StartupCheckUpdate:       true,

		ProgressFirstEmptyChar: []rune(constants.ProgressEmptyChar)[0],
		ProgressEmptyChar:      []rune(constants.ProgressEmptyChar)[0],
		ProgressLastEmptyChar:  []rune(constants.ProgressEmptyChar)[0],
		ProgressFirstFullChar:  []rune(constants.ProgressFullChar)[0],
		ProgressFullChar:       []rune(constants.ProgressFullChar)[0],
		ProgressLastFullChar:   []rune(constants.ProgressFullChar)[0],

		ShowTitle:            true,
		LoadingText:          constants.MainLoadingText,
		PlayerSongLevel:      service.Higher,
		PrimaryColor:         constants.AppPrimaryColor,
		ShowLyric:            true,
		ShowLyricTrans:       true,
		ShowNotify:           true,
		NotifyIcon:           constants.DefaultNotifyIcon,
		PProfPort:            constants.MainPProfPort,
		AltScreen:            true,
		EnableMouseEvent:     true,
		PlayerEngine:         constants.BeepPlayer,
		PlayerBeepMp3Decoder: constants.BeepGoMp3Decoder,
	}

	if runtime.GOOS == "darwin" {
		registry.PlayerEngine = constants.OsxPlayer
	}

	return registry
}

func NewRegistryFromIniFile(filepath string) *Registry {
	registry := NewRegistryWithDefault()

	if err := ini.LoadExists(filepath); err != nil {
		return registry
	}

	registry.StartupShow = ini.Bool("startup.show", true)
	registry.StartupProgressOutBounce = ini.Bool("startup.progressOutBounce", true)
	registry.StartupLoadingDuration = time.Second * time.Duration(ini.Int("startup.loadingSeconds", constants.StartupLoadingSeconds))
	registry.StartupWelcome = ini.String("startup.welcome", constants.AppName)
	registry.StartupCheckUpdate = ini.Bool("startup.checkUpdate", true)

	emptyChar := ini.String("progress.emptyChar", constants.ProgressEmptyChar)
	registry.ProgressEmptyChar = firstCharOrDefault(emptyChar, constants.ProgressEmptyChar)
	firstEmptyChar := ini.String("progress.firstEmptyChar", constants.ProgressEmptyChar)
	registry.ProgressFirstEmptyChar = firstCharOrDefault(firstEmptyChar, constants.ProgressEmptyChar)
	lastEmptyChar := ini.String("progress.lastEmptyChar", constants.ProgressEmptyChar)
	registry.ProgressLastEmptyChar = firstCharOrDefault(lastEmptyChar, constants.ProgressEmptyChar)

	fullChar := ini.String("progress.fullChar", constants.ProgressFullChar)
	registry.ProgressFullChar = firstCharOrDefault(fullChar, constants.ProgressFullChar)
	firstFullChar := ini.String("progress.firstFullChar", constants.ProgressFullChar)
	registry.ProgressFirstFullChar = firstCharOrDefault(firstFullChar, constants.ProgressFullChar)
	lastFullChar := ini.String("progress.lastFullChar", constants.ProgressFullChar)
	registry.ProgressLastFullChar = firstCharOrDefault(lastFullChar, constants.ProgressFullChar)

	registry.SpotifyClientId = ini.Get("main.spotifyClientId", "")
	registry.ShowTitle = ini.Bool("main.showTitle", true)
	registry.LoadingText = ini.String("main.loadingText", constants.MainLoadingText)
	songLevel := service.SongQualityLevel(ini.String("main.songLevel", string(service.Higher)))
	if songLevel.IsValid() {
		registry.PlayerSongLevel = songLevel
	}
	primaryColor := ini.String("main.primaryColor", constants.AppPrimaryColor)
	if primaryColor != "" {
		registry.PrimaryColor = primaryColor
	} else {
		registry.PrimaryColor = constants.AppPrimaryColor
	}
	registry.ShowLyric = ini.Bool("main.showLyric", true)
	registry.LyricOffset = ini.Int("main.lyricOffset", 0)
	registry.ShowLyricTrans = ini.Bool("main.showLyricTrans", true)
	registry.ShowNotify = ini.Bool("main.showNotify", true)
	registry.NotifyIcon = ini.String("main.notifyIcon", constants.DefaultNotifyIcon)
	registry.PProfPort = ini.Int("main.pprofPort", constants.MainPProfPort)
	registry.AltScreen = ini.Bool("main.altScreen", true)
	registry.EnableMouseEvent = ini.Bool("main.enableMouseEvent", true)
	registry.DoubleColumn = ini.Bool("main.doubleColumn", true)

	defaultPlayer := constants.BeepPlayer
	if runtime.GOOS == "darwin" {
		defaultPlayer = constants.OsxPlayer
	}
	registry.PlayerEngine = ini.String("player.engine", defaultPlayer)
	registry.PlayerBeepMp3Decoder = ini.String("player.beepMp3Decoder", constants.BeepGoMp3Decoder)

	return registry
}

func firstCharOrDefault(s, defaultStr string) rune {
	if len(s) > 0 {
		return []rune(s)[0]
	}
	return []rune(defaultStr)[0]
}
