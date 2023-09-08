package ui

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/arcspace/go-arc-sdk/stdlib/task"
	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	"github.com/arcspace/go-librespot/librespot/core"
	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/lyric"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"
)

func NewSpotifySession() respot.Session {
	ctx := respot.DefaultSessionContext(types.SpotifyDeviceName)
	sess, err := respot.StartNewSession(ctx)
	if err != nil {
		panic(err)
	}
	if se, ok := sess.(*core.Session); ok {
		se.Downloader().SetAudioFormat(configs.ConfigRegistry.Main.SongFormat.ToSpotifyFormat())
	}
	ctx.Context, _ = task.Start(&task.Task{Label: types.SpotifyDeviceName})
	return sess
}

func (s *Spotifox) ReconnSessionWhenNeed(f func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = f()
		if err == nil {
			return nil
		}
		if s.CheckConnectErr(err) == utils.NeedReconnect {
			s.sess = NewSpotifySession()
		}
	}
	return err
}

func (s *Spotifox) CheckAuthSession() utils.ResCode {
	if s.spotifyClient == nil {
		return utils.NeedLogin
	}
	if s.user == nil || s.user.ID == "" || s.user.Token.AccessToken == "" {
		return utils.NeedLogin
	}

	return utils.Success
}

func (s *Spotifox) CheckConnectErr(err error) utils.ResCode {
	if s.sess == nil || errors.Is(err, io.EOF) {
		return utils.NeedReconnect
	}
	return utils.UnknownError
}

func (s *Spotifox) FetchSongLyrics(songId spotify.ID) *lyric.LRCFile {
	if s.lyricClient == nil {
		lrcFile, _ := lyric.ReadLRC(strings.NewReader("[00:00.00] No Lyrics~ (Please set your cookies in the configuration file)"))
		return lrcFile
	}

	// get by user's cookie
	l, err := s.lyricClient.Get(string(songId))
	if err != nil || l == nil || l.Lyrics == nil {
		utils.Logger().Printf("get song lyrics failed: %+v", err)
		return nil
	}
	var frags []lyric.LRCFragment
	for _, v := range l.Lyrics.Lines {
		frags = append(frags, lyric.LRCFragment{StartTimeMs: int64(v.Time), Content: v.Words})
	}
	return lyric.NewLRCFileFromFrags(frags)
}

func (s *Spotifox) CheckLikedSong(songId spotify.ID) bool {
	if s.spotifyClient == nil {
		return false
	}
	res, err := s.spotifyClient.UserHasTracks(context.Background(), songId)
	if err != nil {
		utils.Logger().Printf("check liked song failed: %+v", err)
		return false
	}
	if len(res) == 0 {
		return false
	}
	return res[0]
}

func (s *Spotifox) LikeSong(songId spotify.ID, likeOrNot bool) bool {
	if s.spotifyClient == nil {
		return false
	}
	var err error
	if likeOrNot {
		err = s.spotifyClient.AddTracksToLibrary(context.Background(), songId)
	} else {
		err = s.spotifyClient.RemoveTracksFromLibrary(context.Background(), songId)
	}
	if err != nil {
		utils.Logger().Printf("Change liked song failed: %+v", err)
		return false
	}
	return true
}

func (s *Spotifox) FollowPlaylist(id spotify.ID, followOrNot bool) bool {
	if s.spotifyClient == nil {
		return false
	}
	var err error
	if followOrNot {
		err = s.spotifyClient.FollowPlaylist(context.Background(), id, true)
	} else {
		err = s.spotifyClient.UnfollowPlaylist(context.Background(), id)
	}
	if err != nil {
		utils.Logger().Printf("Change followed playlist failed: %+v", err)
		return false
	}
	return true
}
