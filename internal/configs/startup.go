package configs

import (
	"github.com/anhoder/foxful-cli/model"
)

type StartupOptions struct {
	model.StartupOptions
	CheckUpdate bool
}
