package storage

import (
	"github.com/go-musicfox/spotifox/pkg/constants"
)

type ExtInfo struct {
	StorageVersion string `json:"storage_version"`
}

func (e ExtInfo) GetDbName() string {
	return constants.AppDBName
}

func (e ExtInfo) GetTableName() string {
	return "default_bucket"
}

func (e ExtInfo) GetKey() string {
	return "ext_info"
}
