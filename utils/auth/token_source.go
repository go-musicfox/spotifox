package auth

import (
	"errors"
	"time"

	"golang.org/x/oauth2"
)

var _ oauth2.TokenSource = (*TokenSourceWrapper)(nil)

type TokenSourceWrapper oauth2.Token

var ErrTokenExpired = errors.New("Token has expired")

func (s *TokenSourceWrapper) Token() (*oauth2.Token, error) {
	if time.Now().After(s.Expiry) {
		return nil, ErrTokenExpired
	}

	return (*oauth2.Token)(s), nil
}
