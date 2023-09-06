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
	NetworkError
	NeedLogin
	NeedReconnect
	PasswordError
)

// CheckCode 验证响应码
func CheckCode(code float64) ResCode {
	switch code {
	case 301, 302, 20001:
		return NeedLogin
	case 520:
		return NetworkError
	case 200:
		return Success
	}

	return PasswordError
}

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

// ReplaceSpecialStr 替换特殊字符
func ReplaceSpecialStr(str string) string {
	return specialCharReplacer.Replace(str)
}
