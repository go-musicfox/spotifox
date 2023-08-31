package storage

import (
	"github.com/go-musicfox/spotifox/pkg/constants"
)

type User struct{}

func (u User) GetDbName() string {
	return constants.AppDBName
}

func (u User) GetTableName() string {
	return "default_bucket"
}

func (u User) GetKey() string {
	return "cur_user"
}
