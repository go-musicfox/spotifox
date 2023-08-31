//go:build !darwin

package entry

import "github.com/go-musicfox/spotifox/utils"

func AppEntry() {
	defer utils.Recover(false)

	runCLI()
}
