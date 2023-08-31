package commands

import (
	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/spotifox/pkg/configs"
	"github.com/go-musicfox/spotifox/pkg/constants"
	"github.com/go-musicfox/spotifox/pkg/ui"
	"github.com/go-musicfox/spotifox/utils"

	"github.com/gookit/gcli/v2"
)

func NewPlayerCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "spotifox",
		UseFor: "Command line player for Spotify",
		Func:   runPlayer,
	}
	return cmd
}

func runPlayer(_ *gcli.Command, _ []string) error {
	if GlobalOptions.PProfMode {
		go utils.PanicRecoverWrapper(true, func() {
			panic(http.ListenAndServe(":"+strconv.Itoa(configs.ConfigRegistry.PProfPort), nil))
		})
	}

	http.DefaultClient.Timeout = constants.AppHttpTimeout

	var opts = model.DefaultOptions()
	configs.ConfigRegistry.FillToModelOpts(opts)

	var (
		spotifox     = ui.NewSpotifox(model.NewApp(opts))
		eventHandler = ui.NewEventHandler(spotifox)
	)
	spotifox.App.With(
		model.WithHook(spotifox.InitHook, spotifox.CloseHook),
		model.WithMainMenu(ui.NewMainMenu(spotifox), &model.MenuItem{Title: "Spotifox"}),
		func(options *model.Options) {
			options.Components = append(options.Components, spotifox.Player())
			options.KBControllers = append(options.KBControllers, eventHandler)
			options.MouseControllers = append(options.MouseControllers, eventHandler)
		},
	)

	return spotifox.Run()
}
