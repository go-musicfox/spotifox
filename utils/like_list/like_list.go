package like_list

import (
	"github.com/zmb3/spotify/v2"
)

type LikeList map[spotify.ID]struct{}

var likeList = make(LikeList)

func IsLikeSong(songId spotify.ID) bool {
	_, ok := likeList[songId]
	return ok
}

func RefreshLikeList(userId string) {
	// s := &service.LikeListService{UID: strconv.FormatInt(userId, 10)}
	// _, resp := s.LikeList()

	// likeList = make(LikeList)
	// _, _ = jsonparser.ArrayEach(resp, func(value []byte, _ jsonparser.ValueType, _ int, err error) {
	// 	if err != nil {
	// 		return
	// 	}
	// 	if id, err := jsonparser.ParseInt(value); err == nil {
	// 		likeList[id] = struct{}{}
	// 	}
	// }, "ids")
}
