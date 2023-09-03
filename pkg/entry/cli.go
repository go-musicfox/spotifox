package entry

import (
	"fmt"
	"log"

	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/spotifox/pkg/commands"
	"github.com/go-musicfox/spotifox/pkg/configs"
	"github.com/go-musicfox/spotifox/pkg/constants"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/gookit/gcli/v2"
)

func runCLI() {
	log.SetOutput(utils.LogWriter())

	var app = gcli.NewApp()
	app.Name = constants.AppName
	app.Version = constants.AppVersion
	app.Description = constants.AppDescription
	app.GOptsBinder = func(gf *gcli.Flags) {
		gf.BoolOpt(&commands.GlobalOptions.PProfMode, "pprof", "p", false, "enable PProf mode")
	}

	utils.LoadIniConfig()

	util.PrimaryColor = configs.ConfigRegistry.PrimaryColor
	var (
		logo         = util.GetAlphaAscii(app.Name)
		randomColor  = util.GetPrimaryColor()
		logoColorful = util.SetFgStyle(logo, randomColor)
	)

	gcli.AppHelpTemplate = fmt.Sprintf(constants.AppHelpTemplate, logoColorful)
	app.Logo.Text = logoColorful

	var playerCommand = commands.NewPlayerCommand()
	app.Add(playerCommand)
	app.Add(commands.NewConfigCommand())
	app.DefaultCommand(playerCommand.Name)

	app.Run()
}
