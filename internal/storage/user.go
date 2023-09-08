package storage

import (
	"github.com/go-musicfox/spotifox/internal/types"
)

type User struct{}

func (u User) GetDbName() string {
	return types.AppDBName
}

func (u User) GetTableName() string {
	return "default_bucket"
}

func (u User) GetKey() string {
	return "cur_user"
}
