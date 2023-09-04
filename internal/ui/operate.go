package ui

import (
	"os"
	"path"
	"strconv"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/spotifox/internal/constants"
	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

func likePlayingSong(m *Spotifox, likeOrNot bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return nil
	}

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			likePlayingSong(m, likeOrNot)
			return nil
		})
		return page
	}

	if !m.LikeSong(m.player.playlist[m.player.curSongIndex].ID, likeOrNot) {
		return nil
	}
	m.player.isCurSongLiked = likeOrNot

	var title = "已添加到我喜欢的歌曲"
	if !likeOrNot {
		title = "已从我喜欢的歌曲移除"
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    m.player.playlist[m.player.curSongIndex].Name,
		Url:     utils.WebURLOfLibrary(),
		GroupId: constants.GroupID,
	})
	return nil
}

func logout(clearAll bool) {
	table := storage.NewTable()
	_ = table.DeleteByKVModel(storage.User{})
	if clearAll {
		(&storage.LastfmUser{}).Clear()
	}
	utils.Notify(utils.NotifyContent{
		Title:   "登出成功",
		Text:    "已清理用户信息",
		Url:     constants.AppGithubUrl,
		GroupId: constants.GroupID,
	})
	_ = os.Remove(path.Join(utils.GetLocalDataDir(), "cookie"))
}

func likeSelectedSong(m *Spotifox, likeOrNot bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(SongsMenu)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return nil
	}
	songs := me.Songs()

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			likeSelectedSong(m, likeOrNot)
			return nil
		})
		return page
	}

	if !m.LikeSong(songs[selectedIndex].ID, likeOrNot) {
		return nil
	}

	var title = "已添加到我喜欢的歌曲"
	if !likeOrNot {
		title = "已从我喜欢的歌曲移除"
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    songs[selectedIndex].Name,
		Url:     utils.WebURLOfLibrary(),
		GroupId: constants.GroupID,
	})
	return nil
}

func albumOfPlayingSong(m *Spotifox) {
	// loading := model.NewLoading(m.MustMain())
	// loading.Start()
	// defer loading.Complete()

	// var (
	// 	main = m.MustMain()
	// 	menu = main.CurMenu()
	// )
	// if m.player.curSongIndex >= len(m.player.playlist) {
	// 	return
	// }

	// curSong := m.player.playlist[m.player.curSongIndex]
	// 避免重复进入
	// if detail, ok := menu.(*AlbumDetailMenu); ok && detail.albumId == curSong.Album.Id {
	// 	return
	// }

	// main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(m), curSong.Album.Id), &model.MenuItem{Title: curSong.Album.Name, Subtitle: "「" + curSong.Name + "」所属专辑"})
}

func albumOfSelectedSong(m *Spotifox) {
	// loading := model.NewLoading(m.MustMain())
	// loading.Start()
	// defer loading.Complete()

	// var (
	// 	main = m.MustMain()
	// 	menu = main.CurMenu()
	// )
	// me, ok := menu.(SongsMenu)
	// selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	// if !ok || selectedIndex >= len(me.Songs()) {
	// 	return
	// }
	// songs := me.Songs()

	// // 避免重复进入
	// if detail, ok := menu.(*AlbumDetailMenu); ok && detail.albumId == songs[selectedIndex].Album.Id {
	// 	return
	// }

	// main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(m), songs[selectedIndex].Album.Id), &model.MenuItem{Title: songs[selectedIndex].Album.Name, Subtitle: "「" + songs[selectedIndex].Name + "」所属专辑"})
}

func artistOfPlayingSong(m *Spotifox) {
	// loading := model.NewLoading(m.MustMain())
	// loading.Start()
	// defer loading.Complete()

	// var (
	// 	main = m.MustMain()
	// 	menu = main.CurMenu()
	// )
	// if m.player.curSongIndex >= len(m.player.playlist) {
	// 	return
	// }
	// curSong := m.player.playlist[m.player.curSongIndex]
	// artistCount := len(curSong.Artists)
	// if artistCount <= 0 {
	// 	return
	// }
	// if artistCount == 1 {
	// 	// 避免重复进入
	// 	if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == curSong.Artists[0].Id {
	// 		return
	// 	}
	// 	main.EnterMenu(NewArtistDetailMenu(newBaseMenu(m), curSong.Artists[0].Id, curSong.Artists[0].Name), &model.MenuItem{Title: curSong.Artists[0].Name, Subtitle: "「" + curSong.Name + "」所属歌手"})
	// 	return
	// }
	// // 避免重复进入
	// if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.Id == curSong.Id {
	// 	return
	// }
	// main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(m), curSong), &model.MenuItem{Title: "「" + curSong.Name + "」所属歌手"})
}

func artistOfSelectedSong(m *Spotifox) {
	// loading := model.NewLoading(m.MustMain())
	// loading.Start()
	// defer loading.Complete()

	// var (
	// 	main = m.MustMain()
	// 	menu = main.CurMenu()
	// )
	// me, ok := menu.(SongsMenu)
	// selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	// if !ok || selectedIndex >= len(me.Songs()) {
	// 	return
	// }
	// songs := me.Songs()
	// song := songs[selectedIndex]
	// artistCount := len(song.Artists)
	// if artistCount <= 0 {
	// 	return
	// }
	// if artistCount == 1 {
	// 	// 避免重复进入
	// 	if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == song.Artists[0].Id {
	// 		return
	// 	}
	// 	main.EnterMenu(NewArtistDetailMenu(newBaseMenu(m), song.Artists[0].Id, song.Artists[0].Name), &model.MenuItem{Title: song.Artists[0].Name, Subtitle: "「" + song.Name + "」所属歌手"})
	// 	return
	// }
	// // 避免重复进入
	// if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.Id == song.Id {
	// 	return
	// }
	// main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(m), song), &model.MenuItem{Title: "「" + song.Name + "」所属歌手"})
}

func openPlayingSongInWeb(m *Spotifox) {
	// loading := model.NewLoading(m.MustMain())
	// loading.Start()
	// defer loading.Complete()

	// if m.player.curSongIndex >= len(m.player.playlist) {
	// 	return
	// }
	// curSong := m.player.playlist[m.player.curSongIndex]

	// _ = open.Start(utils.WebUrlOfSong(curSong.Id))
}

func openSelectedItemInWeb(m *Spotifox) {
	// loading := model.NewLoading(m.MustMain())
	// loading.Start()
	// defer loading.Complete()

	// var (
	// 	main = m.MustMain()
	// 	menu = main.CurMenu()
	// )
	// selectedIndex := menu.RealDataIndex(main.SelectedIndex())

	// // 打开歌曲
	// if songMenu, ok := menu.(SongsMenu); ok && selectedIndex < len(songMenu.Songs()) {
	// 	_ = open.Start(utils.WebUrlOfSong(songMenu.Songs()[selectedIndex].Id))
	// 	return
	// }

	// // 打开歌单
	// if playlistMenu, ok := menu.(PlaylistsMenu); ok && selectedIndex < len(playlistMenu.Playlists()) {
	// 	_ = open.Start(utils.WebUrlOfPlaylist(playlistMenu.Playlists()[selectedIndex].Id))
	// 	return
	// }

	// // 打开专辑
	// if albumMenu, ok := menu.(AlbumsMenu); ok && selectedIndex < len(albumMenu.Albums()) {
	// 	_ = open.Start(utils.WebUrlOfAlbum(albumMenu.Albums()[selectedIndex].Id))
	// 	return
	// }

	// // 打开歌手
	// if artistMenu, ok := menu.(ArtistsMenu); ok && selectedIndex < len(artistMenu.Artists()) {
	// 	_ = open.Start(utils.WebUrlOfArtist(artistMenu.Artists()[selectedIndex].Id))
	// 	return
	// }
}

func followSelectedPlaylist(m *Spotifox, followOrNot bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			followSelectedPlaylist(m, followOrNot)
			return nil
		})
		return page
	}

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(PlaylistsMenu)
	if !ok || main.SelectedIndex() >= len(me.Playlists()) {
		return nil
	}
	playlists := me.Playlists()

	if !m.FollowPlaylist(playlists[main.SelectedIndex()].ID, followOrNot) {
		return nil
	}

	var title = "已关注歌单"
	if !followOrNot {
		title = "已取消关注歌单"
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    playlists[main.SelectedIndex()].Name,
		Url:     constants.AppGithubUrl,
		GroupId: constants.GroupID,
	})
	return nil
}

func openAddSongToUserPlaylistMenu(m *Spotifox, isSelected, isAdd bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			openAddSongToUserPlaylistMenu(m, isSelected, isAdd)
			return nil
		})
		return page
	}

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	switch me := menu.(type) {
	case SongsMenu:
		if menu.RealDataIndex(main.SelectedIndex()) >= len(me.Songs()) {
			return nil
		}
	default:
		if isSelected {
			return nil
		}
	}
	// 避免重复进入
	if _, ok := menu.(*AddToUserPlaylistMenu); ok {
		return nil
	}
	//var song structs.Song
	//var subtitle string
	//if isSelected {
	//	song = menu.(SongsMenu).Songs()[menu.RealDataIndex(main.SelectedIndex())]
	//} else {
	//	song = m.player.curSong
	//}
	//if isAdd {
	//	subtitle = "将「" + song.Name + "」加入歌单"
	//} else {
	//	subtitle = "将「" + song.Name + "」从歌单中删除"
	//}
	//main.EnterMenu(NewAddToUserPlaylistMenu(newBaseMenu(m), m.user.UserId, song, isAdd), &model.MenuItem{Title: "我的歌单", Subtitle: subtitle})
	return nil
}

func addSongToUserPlaylist(m *Spotifox, isAdd bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			addSongToUserPlaylist(m, isAdd)
			return nil
		})
		return page
	}

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me := menu.(*AddToUserPlaylistMenu)
	playlist := me.playlists[menu.RealDataIndex(main.SelectedIndex())]

	var op string
	if isAdd {
		op = "add"
	} else {
		op = "del"
	}
	likeService := service.PlaylistTracksService{
		TrackIds: []string{strconv.FormatInt(me.song.Id, 10)},
		Op:       op,
		Pid:      string(playlist.ID),
	}
	if code, resp := likeService.PlaylistTracks(); code != 200 {
		var msg string
		if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
			msg, _ = jsonparser.GetString(resp, "data", "message")
		}
		if msg == "" && isAdd {
			msg = "加入歌单失败"
		} else if msg == "" && !isAdd {
			msg = "从歌单中删除失败"
		}
		utils.Notify(utils.NotifyContent{
			Title:   msg,
			Text:    me.song.Name,
			Url:     constants.AppGithubUrl,
			GroupId: constants.GroupID,
		})
		main.BackMenu()
		return nil
	}

	// var title string
	// if isAdd {
	// 	title = "已添加到歌单「" + playlist.Name + "」"
	// } else {
	// 	title = "已从歌单「" + playlist.Name + "」中删除"
	// }
	// utils.Notify(utils.NotifyContent{
	// 	Title:   title,
	// 	Text:    me.song.Name,
	// 	Url:     utils.WebUrlOfPlaylist(playlist.ID),
	// 	GroupId: constants.GroupID,
	// })
	main.BackMenu()
	switch mt := menu.(type) {
	case *PlaylistDetailMenu:
		// 刷新菜单
		if !isAdd && mt.playlistId == playlist.ID {
			t := main.MenuTitle()
			main.BackMenu()
			_, page := menu.BeforeEnterMenuHook()(main)
			main.EnterMenu(menu, t)
			return page
		}
	default:
	}
	return nil
}
