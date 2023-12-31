package entry

import (
	"fmt"
	"log"

	"github.com/anhoder/foxful-cli/util"
	"github.com/go-musicfox/spotifox/internal/commands"
	"github.com/go-musicfox/spotifox/internal/configs"
	"github.com/go-musicfox/spotifox/internal/types"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/gookit/gcli/v2"
)

func runCLI() {
	log.SetOutput(utils.LogWriter())

	var app = gcli.NewApp()
	app.Name = types.AppName
	app.Version = types.AppVersion
	app.Description = types.AppDescription
	app.GOptsBinder = func(gf *gcli.Flags) {
		gf.BoolOpt(&commands.GlobalOptions.PProfMode, "pprof", "p", false, "enable PProf mode")
	}

	utils.LoadIniConfig()

	util.PrimaryColor = configs.ConfigRegistry.Main.PrimaryColor
	var (
		logo         = util.GetAlphaAscii(app.Name)
		randomColor  = util.GetPrimaryColor()
		logoColorful = util.SetFgStyle(logo, randomColor)
	)

	gcli.AppHelpTemplate = fmt.Sprintf(types.AppHelpTemplate, logoColorful)
	app.Logo.Text = logoColorful

	var playerCommand = commands.NewPlayerCommand()
	app.Add(playerCommand)
	app.Add(commands.NewConfigCommand())
	app.DefaultCommand(playerCommand.Name)

	app.Run()
}
