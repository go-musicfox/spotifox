package ui

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/go-musicfox/spotifox/internal/lastfm"
	"github.com/go-musicfox/spotifox/internal/lyric"
	"github.com/go-musicfox/spotifox/internal/player"
	"github.com/go-musicfox/spotifox/internal/state_handler"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/like_list"
	"github.com/zmb3/spotify/v2"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

// PlayDirection 下首歌的方向
type PlayDirection uint8

const (
	DurationNext PlayDirection = iota
	DurationPrev
)

type CtrlType string

type CtrlSignal struct {
	Type     CtrlType
	Duration time.Duration
}

const (
	CtrlResume   CtrlType = "Resume"
	CtrlPaused   CtrlType = "Paused"
	CtrlStop     CtrlType = "Stop"
	CtrlToggle   CtrlType = "Toggle"
	CtrlPrevious CtrlType = "Previous"
	CtrlNext     CtrlType = "Next"
	CtrlSeek     CtrlType = "Seek"
	CtrlRerender CtrlType = "Rerender"
)

// Player 网易云音乐播放器
type Player struct {
	spotifox *Spotifox
	cancel   context.CancelFunc

	playlist         []spotify.FullTrack // 歌曲列表
	playlistUpdateAt time.Time           // 播放列表更新时间
	curSongIndex     int                 // 当前歌曲的下标
	curSong          spotify.FullTrack   // 当前歌曲信息（防止播放列表发生变动后，歌曲信息不匹配）
	playingMenuKey   string              // 正在播放的菜单Key
	playingMenu      Menu
	playedTime       time.Duration // 已经播放的时长

	lrcTimer          *lyric.LRCTimer   // 歌词计时器
	lyrics            [5]string         // 歌词信息，保留5行
	showLyric         bool              // 显示歌词
	lyricStartRow     int               // 歌词开始行
	lyricLines        int               // 歌词显示行数，3或5
	lyricNowScrollBar *utils.XScrollBar // 当前歌词滚动

	// 播放进度条
	progressLastWidth float64
	progressRamp      []string

	playErrCount int // 错误计数，当错误连续超过5次，停止播放
	mode         player.Mode
	stateHandler *state_handler.Handler
	ctrl         chan CtrlSignal

	player.Player // 播放器
}

func NewPlayer(spotifox *Spotifox) *Player {
	p := &Player{
		spotifox:          spotifox,
		mode:              player.PmListLoop,
		ctrl:              make(chan CtrlSignal),
		lyricNowScrollBar: utils.NewXScrollBar(),
	}
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	p.Player = player.NewPlayerFromConfig()
	p.stateHandler = state_handler.NewHandler(p, p.PlayingInfo())

	// remote control
	go utils.PanicRecoverWrapper(false, func() {
		for {
			select {
			case <-ctx.Done():
				return
			case signal := <-p.ctrl:
				p.handleControlSignal(signal)
			}
		}
	})

	// 状态监听
	go utils.PanicRecoverWrapper(false, func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-p.Player.StateChan():
				p.stateHandler.SetPlayingInfo(p.PlayingInfo())
				if s != player.Stopped {
					p.spotifox.Rerender(false)
					break
				}
				// report to lastfm
				lastfm.Report(p.spotifox.lastfm, lastfm.ReportPhaseComplete, p.curSong, p.PassedTime())
				_ = p.NextSong(false)
			}
		}
	})

	// 时间监听
	go utils.PanicRecoverWrapper(false, func() {
		for {
			select {
			case <-ctx.Done():
				return
			case duration := <-p.TimeChan():
				// 200ms 为刷新间隔，刷新间隔修改时此处需要保持同步
				p.playedTime += time.Millisecond * 200
				if duration.Seconds()-p.CurMusic().Duration().Seconds() > 10 {
					// 上报
					lastfm.Report(p.spotifox.lastfm, lastfm.ReportPhaseComplete, p.curSong, p.PassedTime())
					_ = p.NextSong(false)
				}
				if p.lrcTimer != nil {
					select {
					case p.lrcTimer.Timer() <- duration + time.Millisecond*time.Duration(configs.ConfigRegistry.LyricOffset):
					default:
					}
				}

				p.spotifox.Rerender(false)
			}
		}
	})

	return p
}

func (p *Player) Update(_ tea.Msg, _ *model.App) {
	var main = p.spotifox.MustMain()
	spaceHeight := p.spotifox.WindowHeight() - 5 - main.MenuBottomRow()
	if spaceHeight < 4 || !configs.ConfigRegistry.ShowLyric {
		p.showLyric = false
	} else {
		p.showLyric = true
		if spaceHeight > 6 {
			p.lyricStartRow = (p.spotifox.WindowHeight()-3+main.MenuBottomRow())/2 - 3
			p.lyricLines = 5
		} else {
			p.lyricStartRow = (p.spotifox.WindowHeight()-3+main.MenuBottomRow())/2 - 2
			p.lyricLines = 3
		}
	}
}

func (p *Player) View(a *model.App, main *model.Main) (view string, lines int) {
	var playerBuilder strings.Builder
	playerBuilder.WriteString(p.lyricView())
	playerBuilder.WriteString(p.songView())
	playerBuilder.WriteString("\n\n")
	playerBuilder.WriteString(p.progressView())
	return playerBuilder.String(), a.WindowHeight() - main.SearchBarBottomRow()
}

func (p *Player) lyricView() string {
	var (
		endRow = p.spotifox.WindowHeight() - 4
		main   = p.spotifox.MustMain()
	)

	if !p.showLyric {
		if endRow-main.SearchBarBottomRow() > 0 {
			return strings.Repeat("\n", endRow-main.SearchBarBottomRow())
		} else {
			return ""
		}
	}

	var lyricBuilder strings.Builder
	if p.lyricStartRow > main.SearchBarBottomRow() {
		lyricBuilder.WriteString(strings.Repeat("\n", p.lyricStartRow-main.SearchBarBottomRow()))
	}

	var startCol int
	if main.IsDualColumn() {
		startCol = main.MenuStartColumn() + 3
	} else {
		startCol = main.MenuStartColumn() - 4
	}

	maxLen := p.spotifox.WindowWidth() - startCol - 4
	switch p.lyricLines {
	// 3行歌词
	case 3:
		for i := 1; i <= 3; i++ {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := p.lyricNowScrollBar.Tick(maxLen, p.lyrics[i])
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}

			lyricBuilder.WriteString("\n")
		}
	// 5行歌词
	case 5:
		for i := 0; i < 5; i++ {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := p.lyricNowScrollBar.Tick(maxLen, p.lyrics[i])
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}
			lyricBuilder.WriteString("\n")
		}
	}

	if endRow-p.lyricStartRow-p.lyricLines > 0 {
		lyricBuilder.WriteString(strings.Repeat("\n", endRow-p.lyricStartRow-p.lyricLines))
	}

	return lyricBuilder.String()
}

// songView 歌曲信息UI
func (p *Player) songView() string {
	var (
		builder strings.Builder
		main    = p.spotifox.MustMain()
	)

	var prefixLen = 10
	if main.MenuStartColumn()-4 > 0 {
		prefixLen += 12
		builder.WriteString(strings.Repeat(" ", main.MenuStartColumn()-4))
		builder.WriteString(util.SetFgStyle(fmt.Sprintf("[%s] ", player.ModeName(p.mode)), termenv.ANSIBrightMagenta))
		builder.WriteString(util.SetFgStyle(fmt.Sprintf("%d%% ", p.Volume()), termenv.ANSIBrightBlue))
	}
	if p.State() == player.Playing {
		builder.WriteString(util.SetFgStyle("♫ ♪ ♫ ♪ ", termenv.ANSIBrightYellow))
	} else {
		builder.WriteString(util.SetFgStyle("_ z Z Z ", termenv.ANSIYellow))
	}

	songId := p.curSong.ID
	if songId != "" {
		if like_list.IsLikeSong(songId) {
			builder.WriteString(util.SetFgStyle("♥ ", termenv.ANSIRed))
		} else {
			builder.WriteString(util.SetFgStyle("♥ ", termenv.ANSIWhite))
		}
	}

	if p.curSongIndex < len(p.playlist) {
		// 按剩余长度截断字符串
		truncateSong := runewidth.Truncate(p.curSong.Name, p.spotifox.WindowWidth()-main.MenuStartColumn()-prefixLen, "") // 多减，避免剩余1个中文字符
		builder.WriteString(util.SetFgStyle(truncateSong, util.GetPrimaryColor()))
		builder.WriteString(" ")

		var artists strings.Builder
		for i, v := range utils.ArtistNamesOfSong(&p.curSong) {
			if i != 0 {
				artists.WriteString(",")
			}
			artists.WriteString(v)
		}

		// 按剩余长度截断字符串
		remainLen := p.spotifox.WindowWidth() - main.MenuStartColumn() - prefixLen - runewidth.StringWidth(p.curSong.Name)
		truncateArtists := runewidth.Truncate(
			runewidth.FillRight(artists.String(), remainLen),
			remainLen, "")
		builder.WriteString(util.SetFgStyle(truncateArtists, termenv.ANSIBrightBlack))
	}

	return builder.String()
}

// progressView 进度条UI
func (p *Player) progressView() string {
	allDuration := int(p.CurMusic().Duration().Seconds())
	if allDuration == 0 {
		return ""
	}
	passedDuration := int(p.PassedTime().Seconds())
	progress := passedDuration * 100 / allDuration

	width := float64(p.spotifox.WindowWidth() - 14)
	start, end := model.GetProgressColor()
	if width != p.progressLastWidth || len(p.progressRamp) == 0 {
		p.progressRamp = util.MakeRamp(start, end, width)
		p.progressLastWidth = width
	}

	progressView := model.Progress(&p.spotifox.Options().ProgressOptions, int(width), int(math.Round(width*float64(progress)/100)), p.progressRamp)

	if allDuration/60 >= 100 {
		times := util.SetFgStyle(fmt.Sprintf("%03d:%02d/%03d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), util.GetPrimaryColor())
		return progressView + " " + times
	} else {
		times := util.SetFgStyle(fmt.Sprintf("%02d:%02d/%02d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), util.GetPrimaryColor())
		return progressView + " " + times + " "
	}

}

// InPlayingMenu 是否处于正在播放的菜单中
func (p *Player) InPlayingMenu() bool {
	var key = p.spotifox.MustMain().CurMenu().GetMenuKey()
	return key == p.playingMenuKey || key == CurPlaylistKey
}

// CompareWithCurPlaylist 与当前播放列表对比，是否一致
func (p *Player) CompareWithCurPlaylist(playlist []spotify.FullTrack) bool {
	if len(playlist) != len(p.playlist) {
		return false
	}

	// 如果前20个一致，则认为相同
	for i := 0; i < 20 && i < len(playlist); i++ {
		if !utils.CompareSong(playlist[i], p.playlist[i]) {
			return false
		}
	}

	return true
}

// LocatePlayingSong 定位到正在播放的音乐
func (p *Player) LocatePlayingSong() {
	var (
		main        = p.spotifox.MustMain()
		curMenu, ok = main.CurMenu().(Menu)
	)
	if !ok {
		return
	}

	if !curMenu.IsLocatable() {
		return
	}

	menu, ok := curMenu.(SongsMenu)
	if !ok {
		return
	}
	if !p.InPlayingMenu() || !p.CompareWithCurPlaylist(menu.Songs()) {
		return
	}

	var pageDelta = p.curSongIndex/main.PageSize() - (main.CurPage() - 1)
	if pageDelta > 0 {
		for i := 0; i < pageDelta; i++ {
			p.spotifox.MustMain().NextPage()
		}
	} else if pageDelta < 0 {
		for i := 0; i > pageDelta; i-- {
			p.spotifox.MustMain().PrePage()
		}
	}
	main.SetSelectedIndex(p.curSongIndex)
}

// PlaySong 播放歌曲
func (p *Player) PlaySong(song spotify.FullTrack, direction PlayDirection) model.Page {
	if p.spotifox.CheckSession() == utils.NeedLogin {
		page, _ := p.spotifox.ToLoginPage(func() model.Page {
			p.PlaySong(song, direction)
			return nil
		})
		return page
	}

	loading := model.NewLoading(p.spotifox.MustMain())
	loading.Start()
	defer loading.Complete()

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayerSnapshot{}, storage.PlayerSnapshot{
		CurSongIndex:     p.curSongIndex,
		Playlist:         p.playlist,
		PlaylistUpdateAt: p.playlistUpdateAt,
	})
	p.curSong = song
	p.playedTime = 0

	p.LocatePlayingSong()
	p.Player.Paused()

	asset, err := p.spotifox.sess.PinTrack(string(song.ID), respot.PinOpts{StartInternally: true})
	if err != nil {
		utils.Logger().Printf("spotify pin track err: %+v", err)
		p.progressRamp = []string{}
		p.playErrCount++
		if p.playErrCount >= 3 {
			return nil
		}
		switch direction {
		case DurationPrev:
			return p.PreviousSong(false)
		case DurationNext:
			return p.NextSong(false)
		}
		return nil
	}

	if configs.ConfigRegistry.ShowLyric {
		go p.updateLyric(song.ID)
	}

	p.Player.Play(player.MediaAsset{
		MediaAsset: asset,
		SongInfo:   song,
	})

	// 上报
	lastfm.Report(p.spotifox.lastfm, lastfm.ReportPhaseStart, p.curSong, p.PassedTime())

	go utils.Notify(utils.NotifyContent{
		Title:   "正在播放: " + song.Name,
		Text:    fmt.Sprintf("%s - %s", utils.ArtistNamesOfSong(&song), song.Album.Name),
		Icon:    song.PreviewURL,
		Url:     utils.WebUrlOfSong(song.ID),
		GroupId: constants.GroupID,
	})

	p.playErrCount = 0

	return nil
}

func (p *Player) NextSong(isManual bool) model.Page {
	if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
		var main = p.spotifox.MustMain()
		if p.InPlayingMenu() {
			if main.IsDualColumn() && p.curSongIndex%2 == 0 {
				p.spotifox.MustMain().MoveRight()
			} else {
				p.spotifox.MustMain().MoveDown()
			}
		} else if p.playingMenu != nil {
			if bottomHook := p.playingMenu.BottomOutHook(); bottomHook != nil {
				bottomHook(main)
			}
		}
	}

	switch p.mode {
	case player.PmListLoop:
		p.curSongIndex++
		if p.curSongIndex > len(p.playlist)-1 {
			p.curSongIndex = 0
		}
	case player.PmSingleLoop:
		if isManual && p.curSongIndex < len(p.playlist)-1 {
			p.curSongIndex++
		} else if isManual && p.curSongIndex >= len(p.playlist)-1 {
			return nil
		}
		// else pass
	case player.PmRandom:
		if len(p.playlist)-1 < 0 {
			return nil
		}
		if len(p.playlist)-1 == 0 {
			p.curSongIndex = 0
		} else {
			p.curSongIndex = rand.Intn(len(p.playlist) - 1)
		}
	case player.PmOrder:
		if p.curSongIndex >= len(p.playlist)-1 {
			return nil
		}
		p.curSongIndex++
	}

	if p.curSongIndex > len(p.playlist)-1 {
		return nil
	}
	song := p.playlist[p.curSongIndex]
	return p.PlaySong(song, DurationNext)
}

// PreviousSong 上一曲
func (p *Player) PreviousSong(isManual bool) model.Page {
	if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
		var main = p.spotifox.MustMain()
		if p.InPlayingMenu() {
			if main.IsDualColumn() && p.curSongIndex%2 == 0 {
				p.spotifox.MustMain().MoveUp()
			} else {
				p.spotifox.MustMain().MoveLeft()
			}
		} else if p.playingMenu != nil {
			if topHook := p.playingMenu.TopOutHook(); topHook != nil {
				topHook(main)
			}
		}
	}

	switch p.mode {
	case player.PmListLoop:
		p.curSongIndex--
		if p.curSongIndex < 0 {
			p.curSongIndex = len(p.playlist) - 1
		}
	case player.PmSingleLoop:
		if isManual && p.curSongIndex > 0 {
			p.curSongIndex--
		} else if isManual && p.curSongIndex <= 0 {
			return nil
		}
		// else pass
	case player.PmRandom:
		if len(p.playlist)-1 < 0 {
			return nil
		}
		if len(p.playlist) == 0 {
			p.curSongIndex = 0
		} else {
			p.curSongIndex = rand.Intn(len(p.playlist) - 1)
		}
	case player.PmOrder:
		if p.curSongIndex <= 0 {
			return nil
		}
		p.curSongIndex--
	}

	if p.curSongIndex < 0 {
		return nil
	}
	song := p.playlist[p.curSongIndex]
	return p.PlaySong(song, DurationPrev)
}

func (p *Player) Seek(duration time.Duration) {
	p.Player.Seek(duration)
	if p.lrcTimer != nil {
		p.lrcTimer.Rewind()
	}
	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

// SetPlayMode 播放模式切换
func (p *Player) SetPlayMode(playMode player.Mode) {
	if playMode > 0 {
		p.mode = playMode
	} else {
		switch p.mode {
		case player.PmListLoop, player.PmOrder, player.PmSingleLoop:
			p.mode++
		case player.PmRandom:
			p.mode = player.PmListLoop
		default:
			p.mode = player.PmListLoop
		}
	}

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayMode{}, p.mode)
}

// Close 关闭
func (p *Player) Close() {
	p.cancel()
	if p.stateHandler != nil {
		p.stateHandler.Release()
	}
	p.Player.Close()
}

// lyricListener 歌词变更监听
func (p *Player) lyricListener(_ int64, content, transContent string, _ bool, index int) {
	curIndex := len(p.lyrics) / 2

	// before
	for i := 0; i < curIndex; i++ {
		if f, tf := p.lrcTimer.GetLRCFragment(index - curIndex + i); f != nil {
			p.lyrics[i] = f.Content
			if tf != nil && tf.Content != "" {
				p.lyrics[i] += " [" + tf.Content + "]"
			}
		} else {
			p.lyrics[i] = ""
		}
	}

	// cur
	p.lyrics[curIndex] = content
	if transContent != "" {
		p.lyrics[curIndex] += " [" + transContent + "]"
	}

	// after
	for i := 1; i < len(p.lyrics)-curIndex; i++ {
		if f, tf := p.lrcTimer.GetLRCFragment(index + i); f != nil {
			p.lyrics[curIndex+i] = f.Content
			if tf != nil && tf.Content != "" {
				p.lyrics[curIndex+i] += " [" + tf.Content + "]"
			}
		} else {
			p.lyrics[curIndex+i] = ""
		}
	}
}

func (p *Player) updateLyric(songId spotify.ID) {
	p.lyrics = [5]string{}
	if p.lrcTimer != nil {
		p.lrcTimer.Stop()
	}
	lrcFile, _ := lyric.ReadLRC(strings.NewReader("[00:00.00] No Lyrics~"))
	tranLRCFile, _ := lyric.ReadTranslateLRC(strings.NewReader("[00:00.00]"))
	defer func() {
		p.lrcTimer = lyric.NewLRCTimer(lrcFile, tranLRCFile)
		p.lrcTimer.AddListener(p.lyricListener)
		p.lrcTimer.Start()
	}()

	if l := p.spotifox.FetchSongLyrics(songId); l != nil {
		lrcFile = l
	}
}

func (p *Player) UpVolume() {
	p.Player.UpVolume()

	if v, ok := p.Player.(storage.VolumeStorable); ok {
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.Volume{}, v.Volume())
	}

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) DownVolume() {
	p.Player.DownVolume()

	if v, ok := p.Player.(storage.VolumeStorable); ok {
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.Volume{}, v.Volume())
	}

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) SetVolume(volume int) {
	p.Player.SetVolume(volume)

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) handleControlSignal(signal CtrlSignal) {
	switch signal.Type {
	case CtrlPaused:
		p.Player.Paused()
	case CtrlResume:
		p.Player.Resume()
	case CtrlStop:
		p.Player.Stop()
	case CtrlToggle:
		p.Player.Toggle()
	case CtrlPrevious:
		_ = p.PreviousSong(true)
	case CtrlNext:
		_ = p.NextSong(true)
	case CtrlSeek:
		p.Seek(signal.Duration)
	case CtrlRerender:
		p.spotifox.Rerender(false)
	}
}

func (p *Player) PlayingInfo() state_handler.PlayingInfo {
	// music := p.curSong
	return state_handler.PlayingInfo{
		// TotalDuration:  music.Duration,
		// PassedDuration: p.PassedTime(),
		// State:          p.State(),
		// Volume:         p.Volume(),
		// TrackID:        music.Id,
		// PicUrl:         music.PicUrl,
		// Name:           music.Name,
		// Album:          music.Album.Name,
		// Artist:         music.ArtistName(),
		// AlbumArtist:    music.Album.ArtistName(),
	}
}
