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
	"github.com/arcspace/go-arc-sdk/apis/arc"
	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/lastfm"
	"github.com/go-musicfox/spotifox/internal/lyric"
	"github.com/go-musicfox/spotifox/internal/player"
	"github.com/go-musicfox/spotifox/internal/state_handler"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/locale"
	"github.com/zmb3/spotify/v2"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

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

type Player struct {
	spotifox *Spotifox
	cancel   context.CancelFunc

	playlist         []spotify.FullTrack
	playlistUpdateAt time.Time
	curSongIndex     int
	curSong          spotify.FullTrack
	isCurSongLiked   bool
	playingMenuKey   string
	playingMenu      Menu
	playedTime       time.Duration

	lrcTimer          *lyric.LRCTimer
	lyrics            [5]string
	showLyric         bool
	lyricStartRow     int
	lyricLines        int
	lyricNowScrollBar *utils.XScrollBar

	progressLastWidth float64
	progressRamp      []string

	playErrCount int
	mode         player.Mode
	stateHandler *state_handler.Handler
	ctrl         chan CtrlSignal

	player.Player
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

	go utils.PanicRecoverWrapper(false, func() {
		for {
			select {
			case <-ctx.Done():
				return
			case duration := <-p.TimeChan():
				p.playedTime += time.Millisecond * 200
				if duration.Seconds()-p.CurMusic().Duration().Seconds() > 10 {
					lastfm.Report(p.spotifox.lastfm, lastfm.ReportPhaseComplete, p.curSong, p.PassedTime())
					_ = p.NextSong(false)
				}
				if p.lrcTimer != nil {
					select {
					case p.lrcTimer.Timer() <- duration + time.Millisecond*time.Duration(configs.ConfigRegistry.Main.LyricOffset):
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
	main := p.spotifox.MustMain()
	spaceHeight := p.spotifox.WindowHeight() - 5 - main.MenuBottomRow()
	if spaceHeight < 3 || !configs.ConfigRegistry.Main.ShowLyric {
		p.showLyric = false
	} else {
		p.showLyric = true
		if spaceHeight >= 5 {
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
	return playerBuilder.String(), a.WindowHeight() - main.MenuBottomRow()
}

func (p *Player) lyricView() string {
	var (
		endRow = p.spotifox.WindowHeight() - 4
		main   = p.spotifox.MustMain()
	)

	if !p.showLyric {
		if endRow-main.MenuBottomRow() > 0 {
			return strings.Repeat("\n", endRow-main.MenuBottomRow())
		} else {
			return ""
		}
	}

	var lyricBuilder strings.Builder
	if p.lyricStartRow > main.MenuBottomRow() {
		lyricBuilder.WriteString(strings.Repeat("\n", p.lyricStartRow-main.MenuBottomRow()))
	}

	var startCol int
	if main.IsDualColumn() {
		startCol = main.MenuStartColumn() + 3
	} else {
		startCol = main.MenuStartColumn() - 4
	}

	maxLen := p.spotifox.WindowWidth() - startCol - 4
	switch p.lyricLines {
	// 3 line
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
	// 5 line
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

func (p *Player) songView() string {
	var (
		builder strings.Builder
		main    = p.spotifox.MustMain()
	)

	prefixLen := 10
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
		if p.isCurSongLiked {
			builder.WriteString(util.SetFgStyle("♥ ", termenv.ANSIRed))
		} else {
			builder.WriteString(util.SetFgStyle("♥ ", termenv.ANSIWhite))
		}
	}

	if p.curSongIndex < len(p.playlist) {
		truncateSong := runewidth.Truncate(p.curSong.Name, p.spotifox.WindowWidth()-main.MenuStartColumn()-prefixLen, "")
		builder.WriteString(util.SetFgStyle(truncateSong, util.GetPrimaryColor()))
		builder.WriteString(" ")

		var artists strings.Builder
		for i, v := range utils.ArtistNamesOfSong(&p.curSong) {
			if i != 0 {
				artists.WriteString(",")
			}
			artists.WriteString(v)
		}

		remainLen := p.spotifox.WindowWidth() - main.MenuStartColumn() - prefixLen - runewidth.StringWidth(p.curSong.Name)
		truncateArtists := runewidth.Truncate(
			runewidth.FillRight(artists.String(), remainLen),
			remainLen, "")
		builder.WriteString(util.SetFgStyle(truncateArtists, termenv.ANSIBrightBlack))
	}

	return builder.String()
}

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

func (p *Player) InPlayingMenu() bool {
	key := p.spotifox.MustMain().CurMenu().GetMenuKey()
	return key == p.playingMenuKey || key == CurPlaylistKey
}

func (p *Player) CompareWithCurPlaylist(playlist []spotify.FullTrack) bool {
	if len(playlist) != len(p.playlist) {
		return false
	}

	for i := 0; i < 20 && i < len(playlist); i++ {
		if !utils.CompareSong(playlist[i], p.playlist[i]) {
			return false
		}
	}

	return true
}

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

	pageDelta := p.curSongIndex/main.PageSize() - (main.CurPage() - 1)
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

func (p *Player) PlaySong(song spotify.FullTrack, direction PlayDirection) model.Page {
	if p.spotifox.CheckAuthSession() == utils.NeedLogin {
		page, _ := p.spotifox.ToLoginPage(func() model.Page {
			p.PlaySong(song, direction)
			return nil
		})
		return page
	}

	loading := model.NewLoading(p.spotifox.MustMain())
	loading.Start()
	defer loading.Complete()

	p.isCurSongLiked = p.spotifox.CheckLikedSong(song.ID)

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayerSnapshot{}, storage.PlayerSnapshot{
		CurSongIndex:     p.curSongIndex,
		Playlist:         p.playlist,
		PlaylistUpdateAt: p.playlistUpdateAt,
		IsCurSongLiked:   p.isCurSongLiked,
	})
	p.curSong = song
	p.playedTime = 0

	p.LocatePlayingSong()
	p.Player.Paused()

	var asset arc.MediaAsset
	err := p.spotifox.ReconnSessionWhenNeed(func() error {
		var err error
		asset, err = p.spotifox.sess.PinTrack(string(song.ID), respot.PinOpts{})
		return err
	})
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

	if configs.ConfigRegistry.Main.ShowLyric {
		go p.updateLyric(song.ID)
	}

	p.Player.Play(player.MediaAsset{
		MediaAsset: asset,
		SongInfo:   song,
	})

	lastfm.Report(p.spotifox.lastfm, lastfm.ReportPhaseStart, p.curSong, p.PassedTime())

	go utils.Notify(utils.NotifyContent{
		Title:   locale.MustT("now_playing", locale.WithTplData(map[string]string{"TrackName": song.Name})),
		Text:    fmt.Sprintf("%s - %s", utils.ArtistNameStrOfSong(&song), song.Album.Name),
		Icon:    utils.PicURLOfSong(&song),
		Url:     utils.WebURLOfSong(song.ID),
		GroupId: types.GroupID,
	})
	p.playErrCount = 0

	return nil
}

func (p *Player) NextSong(isManual bool) model.Page {
	if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
		main := p.spotifox.MustMain()
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

func (p *Player) PreviousSong(isManual bool) model.Page {
	if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
		main := p.spotifox.MustMain()
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

func (p *Player) Close() {
	p.cancel()
	if p.stateHandler != nil {
		p.stateHandler.Release()
	}
	p.Player.Close()
}

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
	return state_handler.PlayingInfo{
		TotalDuration:  p.curSong.TimeDuration(),
		PassedDuration: p.PassedTime(),
		State:          p.State(),
		Volume:         p.Volume(),
		TrackID:        string(p.curSong.ID),
		PicUrl:         utils.PicURLOfSong(&p.curSong),
		Name:           p.curSong.Name,
		Album:          p.curSong.Album.Name,
		Artist:         utils.ArtistNameStrOfSong(&p.curSong),
	}
}
