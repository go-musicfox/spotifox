package structs

import (
	"encoding/json"

	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	"github.com/arcspace/go-librespot/librespot/mercury"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

type User struct {
	spotify.User `json:",inline"`

	Username string        `json:"username"`
	Country  string        `json:"country"`
	AuthBlob []byte        `json:"authBlob"`
	Token    mercury.Token `json:"-"`

	Email     string `json:"email"`
	Product   string `json:"product"`
	Birthdate string `json:"birthdate"`
}

func NewUserFromLocalJson(bytes []byte) (User, error) {
	var user User
	if len(bytes) == 0 {
		return user, errors.New("json is empty")
	}
	err := json.Unmarshal(bytes, &user)
	return user, err
}

func NewUserFromSession(session respot.SessionInfo) User {
	return User{
		Username: session.Username,
		Country:  session.Country,
		AuthBlob: session.AuthBlob,
	}
}
