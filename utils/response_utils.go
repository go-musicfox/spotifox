package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-musicfox/spotifox/utils/auth"
	"github.com/zmb3/spotify/v2"
)

type ResCode uint8

const (
	Success ResCode = iota
	UnknownError
	NeedLogin
	NeedReconnect
)

func CheckSpotifyErr(err error) ResCode {
	if err == nil {
		return Success
	}
	if e, ok := err.(spotify.Error); ok && e.Status == http.StatusUnauthorized {
		return NeedLogin
	}
	if errors.Is(err, auth.ErrTokenExpired) {
		return NeedLogin
	}
	return UnknownError
}

var specialCharReplacer = strings.NewReplacer(`“`, `"`, `”`, `"`, `·`, `.`)

func ReplaceSpecialStr(str string) string {
	return specialCharReplacer.Replace(str)
}
