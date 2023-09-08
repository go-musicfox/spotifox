package ui

import (
	"context"
	"os"
	"path"

	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"
	"github.com/zmb3/spotify/v2"

	"github.com/go-musicfox/spotifox/internal/storage"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/go-musicfox/spotifox/utils/locale"
)

func likePlayingSong(m *Spotifox, likeOrNot bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return nil
	}

	if m.CheckAuthSession() == utils.NeedLogin {
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

	var title = locale.MustT("like_song_success")
	if !likeOrNot {
		title = locale.MustT("dislike_song_success")
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    m.player.playlist[m.player.curSongIndex].Name,
		Url:     utils.WebURLOfLibrary(),
		GroupId: types.GroupID,
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
		Title:   locale.MustT("logout_success"),
		Text:    locale.MustT("cleaned_up_user_info"),
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
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

	if m.CheckAuthSession() == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			likeSelectedSong(m, likeOrNot)
			return nil
		})
		return page
	}

	if !m.LikeSong(songs[selectedIndex].ID, likeOrNot) {
		return nil
	}

	var title = locale.MustT("like_song_success")
	if !likeOrNot {
		title = locale.MustT("dislike_song_success")
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    songs[selectedIndex].Name,
		Url:     utils.WebURLOfLibrary(),
		GroupId: types.GroupID,
	})
	return nil
}

func albumOfPlayingSong(m *Spotifox) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	curSong := m.player.playlist[m.player.curSongIndex]
	if detail, ok := menu.(*AlbumDetailMenu); ok && detail.album.ID == curSong.Album.ID {
		return
	}

	main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(m), curSong.Album), &model.MenuItem{Title: curSong.Album.Name, Subtitle: locale.MustT("album_of_track", locale.WithTplData(map[string]string{"TrackName": curSong.Name}))})
}

func albumOfSelectedSong(m *Spotifox) {
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
		return
	}
	songs := me.Songs()

	if detail, ok := menu.(*AlbumDetailMenu); ok && detail.album.ID == songs[selectedIndex].Album.ID {
		return
	}

	main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(m), songs[selectedIndex].Album), &model.MenuItem{Title: songs[selectedIndex].Album.Name, Subtitle: locale.MustT("album_of_track", locale.WithTplData(map[string]string{"TrackName": songs[selectedIndex].Name}))})
}

func artistOfPlayingSong(m *Spotifox) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}
	curSong := m.player.playlist[m.player.curSongIndex]
	artistCount := len(curSong.Artists)
	if artistCount <= 0 {
		return
	}
	if artistCount == 1 {
		if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == curSong.Artists[0].ID {
			return
		}
		main.EnterMenu(NewArtistDetailMenu(newBaseMenu(m), curSong.Artists[0].ID, curSong.Artists[0].Name), &model.MenuItem{Title: curSong.Artists[0].Name, Subtitle: locale.MustT("artist_of_track", locale.WithTplData(map[string]string{"TrackName": curSong.Name}))})
		return
	}
	if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.ID == curSong.ID {
		return
	}
	main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(m), curSong), &model.MenuItem{Title: locale.MustT("artist_of_track", locale.WithTplData(map[string]string{"TrackName": curSong.Name}))})
}

func artistOfSelectedSong(m *Spotifox) {
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
		return
	}
	songs := me.Songs()
	song := songs[selectedIndex]
	artistCount := len(song.Artists)
	if artistCount <= 0 {
		return
	}
	if artistCount == 1 {
		// 避免重复进入
		if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == song.Artists[0].ID {
			return
		}
		main.EnterMenu(NewArtistDetailMenu(newBaseMenu(m), song.Artists[0].ID, song.Artists[0].Name), &model.MenuItem{Title: song.Artists[0].Name, Subtitle: locale.MustT("artist_of_track", locale.WithTplData(map[string]string{"TrackName": song.Name}))})
		return
	}
	// 避免重复进入
	if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.ID == song.ID {
		return
	}
	main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(m), song), &model.MenuItem{Title: locale.MustT("artist_of_track", locale.WithTplData(map[string]string{"TrackName": song.Name}))})
}

func openPlayingSongInWeb(m *Spotifox) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}
	curSong := m.player.playlist[m.player.curSongIndex]

	_ = open.Start(utils.WebURLOfSong(curSong.ID))
}

func openSelectedItemInWeb(m *Spotifox) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())

	if songMenu, ok := menu.(SongsMenu); ok && selectedIndex < len(songMenu.Songs()) {
		_ = open.Start(utils.WebURLOfSong(songMenu.Songs()[selectedIndex].ID))
		return
	}

	if playlistMenu, ok := menu.(PlaylistsMenu); ok && selectedIndex < len(playlistMenu.Playlists()) {
		_ = open.Start(utils.WebURLOfPlaylist(playlistMenu.Playlists()[selectedIndex].ID))
		return
	}

	if albumMenu, ok := menu.(AlbumsMenu); ok && selectedIndex < len(albumMenu.Albums()) {
		_ = open.Start(utils.WebURLOfAlbum(albumMenu.Albums()[selectedIndex].ID))
		return
	}

	if artistMenu, ok := menu.(ArtistsMenu); ok && selectedIndex < len(artistMenu.Artists()) {
		_ = open.Start(utils.WebURLOfArtist(artistMenu.Artists()[selectedIndex].ID))
		return
	}
}

func followSelectedPlaylist(m *Spotifox, followOrNot bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.CheckAuthSession() == utils.NeedLogin {
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

	var title = locale.MustT("follow_playlist_success")
	if !followOrNot {
		title = locale.MustT("unfollow_playlist_success")
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    playlists[main.SelectedIndex()].Name,
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
	})
	return nil
}

func openAddSongToUserPlaylistMenu(m *Spotifox, isSelected, isAdd bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.CheckAuthSession() == utils.NeedLogin {
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
	var song spotify.FullTrack
	var subtitle string
	if isSelected {
		song = menu.(SongsMenu).Songs()[menu.RealDataIndex(main.SelectedIndex())]
	} else {
		song = m.player.curSong
	}
	if isAdd {
		subtitle = locale.MustT("add_song_to_playlist", locale.WithTplData(map[string]string{"TrackName": song.Name}))
	} else {
		subtitle = locale.MustT("remove_song_to_playlist", locale.WithTplData(map[string]string{"TrackName": song.Name}))
	}
	main.EnterMenu(NewAddToUserPlaylistMenu(newBaseMenu(m), song, isAdd), &model.MenuItem{Title: locale.MustT("my_playlists"), Subtitle: subtitle})
	return nil
}

func addSongToUserPlaylist(m *Spotifox, isAdd bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.CheckAuthSession() == utils.NeedLogin {
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
	if len(me.playlists) == 0 {
		return nil
	}
	playlist := me.playlists[menu.RealDataIndex(main.SelectedIndex())]

	_, err := m.spotifyClient.AddTracksToPlaylist(context.Background(), playlist.ID, me.song.ID)
	if utils.CheckSpotifyErr(err) == utils.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			addSongToUserPlaylist(m, isAdd)
			return nil
		})
		return page
	}
	if err != nil {
		utils.Logger().Printf("add song to playlist failed, err: %+v", err)

		return nil
	}

	var title string
	if isAdd {
		title = locale.MustT("add_song_to_playlist_success", locale.WithTplData(map[string]string{"PlaylistName": playlist.Name}))
	} else {
		title = locale.MustT("remove_song_from_playlist_success", locale.WithTplData(map[string]string{"PlaylistName": playlist.Name}))
	}
	utils.Notify(utils.NotifyContent{
		Title:   title,
		Text:    me.song.Name,
		Url:     utils.WebURLOfPlaylist(playlist.ID),
		GroupId: types.GroupID,
	})
	main.BackMenu()

	// refresh menu
	if mt, ok := menu.(*PlaylistDetailMenu); ok && !isAdd && mt.playlistId == playlist.ID {
		t := main.MenuTitle()
		main.BackMenu()
		_, page := menu.BeforeEnterMenuHook()(main)
		main.EnterMenu(menu, t)
		return page
	}
	return nil
}
