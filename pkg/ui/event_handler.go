package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	playerpkg "github.com/go-musicfox/spotifox/pkg/player"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"
)

type EventHandler struct {
	spotifox *Spotifox
}

func NewEventHandler(netease *Spotifox) *EventHandler {
	return &EventHandler{
		spotifox: netease,
	}
}

func (h *EventHandler) KeyMsgHandle(msg tea.KeyMsg, a *model.App) (bool, model.Page, tea.Cmd) {
	var (
		key    = msg.String()
		player = h.spotifox.player
		main   = a.MustMain()
		menu   = main.CurMenu()
	)
	switch key {
	case "enter":
		return h.enterKeyHandle()
	case "c", "C":
		if _, ok := menu.(*CurPlaylist); !ok {
			var subTitle string
			if !player.playlistUpdateAt.IsZero() {
				subTitle = player.playlistUpdateAt.Format("[更新于2006-01-02 15:04:05]")
			}
			main.EnterMenu(NewCurPlaylist(newBaseMenu(h.spotifox), player.playlist), &model.MenuItem{Title: "当前播放列表", Subtitle: subTitle})
			player.LocatePlayingSong()
		}
	case " ", "　":
		newPage := h.spaceKeyHandle()
		if newPage != nil {
			return true, newPage, func() tea.Msg { return newPage.Msg() }
		}
	case "v":
		player.Seek(player.PassedTime() + time.Second*5)
	case "V":
		player.Seek(player.PassedTime() + time.Second*10)
	case "x":
		player.Seek(player.PassedTime() - time.Second*1)
	case "X":
		player.Seek(player.PassedTime() - time.Second*5)
	case "[", "【":
		newPage := player.PreviousSong(true)
		if newPage != nil {
			return true, newPage, func() tea.Msg { return newPage.Msg() }
		}
	case "]", "】":
		newPage := player.NextSong(true)
		if newPage != nil {
			return true, newPage, func() tea.Msg { return newPage.Msg() }
		}
	case "p":
		player.SetPlayMode(0)
	// case "P":
	// newPage := player.Intelligence(false)
	// return true, newPage, a.Tick(time.Nanosecond)
	case ",", "，":
		newPage := likePlayingSong(h.spotifox, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case ".", "。":
		newPage := likePlayingSong(h.spotifox, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "w", "W":
		logout()
		return true, nil, tea.Quit
	case "-", "−", "ー": // half-width, full-width and katakana
		player.DownVolume()
	case "=", "＝":
		player.UpVolume()
	case "t":
		// trash playing song
		newPage := trashPlayingSong(h.spotifox)
		return true, newPage, a.Tick(time.Nanosecond)
	case "T":
		// trash selected song
		newPage := trashSelectedSong(h.spotifox)
		return true, newPage, a.Tick(time.Nanosecond)
	case "<", "〈", "＜", "《", "«": // half-width, full-width, Japanese, Chinese and French
		// like selected song
		newPage := likeSelectedSong(h.spotifox, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case ">", "〉", "＞", "》", "»":
		// unlike selected song
		newPage := likeSelectedSong(h.spotifox, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "?", "？":
		// 帮助
		main.EnterMenu(NewHelpMenu(newBaseMenu(h.spotifox)), &model.MenuItem{Title: "帮助"})
	case "tab":
		newPage := openAddSongToUserPlaylistMenu(h.spotifox, true, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case "shift+tab":
		newPage := openAddSongToUserPlaylistMenu(h.spotifox, true, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "`":
		newPage := openAddSongToUserPlaylistMenu(h.spotifox, false, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case "~", "～":
		newPage := openAddSongToUserPlaylistMenu(h.spotifox, false, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "a":
		// 当前歌曲所属专辑
		albumOfPlayingSong(h.spotifox)
	case "A":
		// 选中歌曲所属专辑
		albumOfSelectedSong(h.spotifox)
	case "s":
		// 当前歌曲所属歌手
		artistOfPlayingSong(h.spotifox)
	case "S":
		// 选中歌曲所属歌手
		artistOfSelectedSong(h.spotifox)
	case "o":
		// 网页打开当前歌曲
		openPlayingSongInWeb(h.spotifox)
	case "O":
		// 网页打开选中项
		openSelectedItemInWeb(h.spotifox)
	case ";", ":", "：", "；":
		// 收藏选中歌单
		newPage := collectSelectedPlaylist(h.spotifox, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case "'", "\"":
		// 取消收藏选中歌单
		newPage := collectSelectedPlaylist(h.spotifox, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "\\", "、":
		// 从播放列表删除歌曲,仅在当前播放列表界面有效
		newPage := delSongFromPlaylist(h.spotifox)
		return true, newPage, a.Tick(time.Nanosecond)
	case "e":
		// 追加到下一曲播放
		addSongToPlaylist(h.spotifox, true)
	case "E":
		// 追加到播放列表末尾
		addSongToPlaylist(h.spotifox, false)
	case "r", "R":
		// rerender
		return true, main, a.RerenderCmd(true)
	default:
		return false, nil, nil
	}

	return true, nil, nil
}

func (h *EventHandler) enterKeyHandle() (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	loading := model.NewLoading(h.spotifox.MustMain())
	loading.Start()
	defer loading.Complete()

	var menu = h.spotifox.MustMain().CurMenu()
	if _, ok := menu.(*AddToUserPlaylistMenu); ok {
		addSongToUserPlaylist(h.spotifox, menu.(*AddToUserPlaylistMenu).action)
		return true, h.spotifox.MustMain(), h.spotifox.Tick(time.Nanosecond)
	}
	return false, nil, nil
}

func (h *EventHandler) spaceKeyHandle() model.Page {
	var (
		songs         []spotify.FullTrack
		inPlayingMenu = h.spotifox.player.InPlayingMenu()
		main          = h.spotifox.MustMain()
		menu          = main.CurMenu()
		player        = h.spotifox.player
	)
	if me, ok := menu.(SongsMenu); ok {
		songs = me.Songs()
	}

	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if me, ok := menu.(Menu); !ok || !me.IsPlayable() || len(songs) == 0 || selectedIndex > len(songs)-1 {
		if player.curSongIndex > len(player.playlist)-1 {
			return nil
		}
		switch player.State() {
		case playerpkg.Paused:
			h.spotifox.player.Resume()
		case playerpkg.Playing:
			h.spotifox.player.Paused()
		case playerpkg.Stopped:
			return player.PlaySong(player.playlist[player.curSongIndex], DurationNext)
		}
		return nil
	}

	if inPlayingMenu && utils.CompareSong(songs[selectedIndex], player.playlist[player.curSongIndex]) {
		switch player.State() {
		case playerpkg.Paused:
			player.Resume()
		case playerpkg.Playing:
			player.Paused()
		}
		return nil
	}

	player.curSongIndex = selectedIndex
	player.playingMenuKey = menu.GetMenuKey()
	if me, ok := menu.(Menu); ok {
		player.playingMenu = me
	}

	var newPlaylists = make([]spotify.FullTrack, len(songs))
	copy(newPlaylists, songs)
	player.playlist = newPlaylists

	player.playlistUpdateAt = time.Now()
	return player.PlaySong(player.playlist[selectedIndex], DurationNext)
}

func (h *EventHandler) MouseMsgHandle(msg tea.MouseMsg, a *model.App) (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	var (
		player = h.spotifox.player
		main   = a.MustMain()
	)
	switch msg.Type {
	case tea.MouseLeft:
		x, y := msg.X, msg.Y
		w := len(player.progressRamp)
		if y+1 == a.WindowHeight() && x+1 <= len(player.progressRamp) {
			allDuration := int(player.CurMusic().Duration().Seconds())
			if allDuration == 0 {
				return true, main, nil
			}
			duration := float64(x) * player.CurMusic().Duration().Seconds() / float64(w)
			player.Seek(time.Second * time.Duration(duration))
			if player.State() != playerpkg.Playing {
				player.Resume()
			}
		}
	case tea.MouseWheelDown:
		player.DownVolume()
	case tea.MouseWheelUp:
		player.UpVolume()
	}

	return true, main, a.Tick(time.Nanosecond)
}
