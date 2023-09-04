package ui

import (
	"context"
	"strings"

	"github.com/go-musicfox/spotifox/internal/lyric"
	"github.com/go-musicfox/spotifox/utils"
	"github.com/zmb3/spotify/v2"
)

func (s *Spotifox) CheckSession() utils.ResCode {
	if s.spotifyClient == nil {
		return utils.NeedLogin
	}
	return utils.CheckUserInfo(s.user)
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
