package structs

import (
	"encoding/json"

	respot "github.com/arcspace/go-librespot/librespot/api-respot"
	"github.com/pkg/errors"
)

type User struct {
	Username string `json:"username"`
	Account  string `json:"account"`
	Country  string `json:"country"`
	AuthBlob []byte `json:"authBlob"`
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
