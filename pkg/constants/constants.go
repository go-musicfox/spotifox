package constants

import (
	"time"
)

var (
	// AppVersion Inject by -ldflags
	AppVersion   = "v1.0.0"
	LastfmKey    = ""
	LastfmSecret = ""
)

const AppName = "spotifox"
const GroupID = "com.go-musicfox.spotifox"
const SpotifyDeviceName = "Spotifox"
const SpotifyOAuthScopes = "playlist-read-private playlist-read-collaborative playlist-modify-private playlist-modify-public user-top-read user-read-recently-played user-library-modify user-library-read"
const AppDescription = "<cyan>Spotifox - Using Spotify on the Command Line</>"
const AppGithubUrl = "https://github.com/go-musicfox/spotifox"
const AppLatestReleases = "https://github.com/go-musicfox/spotifox/releases/latest"
const AppCheckUpdateUrl = "https://api.github.com/repos/go-musicfox/spotifox/releases/latest"
const LastfmAuthUrl = "https://www.last.fm/api/auth/?api_key=%s&token=%s"
const ProgressFullChar = "#"
const ProgressEmptyChar = "."
const StartupLoadingSeconds = 2
const StartupTickDuration = time.Millisecond * 16

const AppLocalDataDir = "spotifox"
const AppDBName = "spotifox"
const AppIniFile = "spotifox.ini"
const AppPrimaryRandom = "random"
const AppPrimaryColor = "#f90022"
const AppHttpTimeout = time.Second * 5

const MainLoadingText = "[加载中...]"
const MainPProfPort = 9876
const DefaultNotifyIcon = "logo.png"

const BeepPlayer = "beep" // beep
const OsxPlayer = "osx"   // osx

const BeepGoMp3Decoder = "go-mp3"
const BeepMiniMp3Decoder = "minimp3"

const SearchPageSize = 100

const AppHelpTemplate = `%s

{{.Description}} (Version: <info>{{.Version}}</>)

<comment>Usage:</>
  {$binName} [Global Options...] <info>{command}</> [--option ...] [argument ...]

<comment>Global Options:</>
{{.GOpts}}
<comment>Available Commands:</>{{range $module, $cs := .Cs}}{{if $module}}
<comment> {{ $module }}</>{{end}}{{ range $cs }}
  <info>{{.Name | paddingName }}</> {{.UseFor}}{{if .Aliases}} (alias: <cyan>{{ join .Aliases ","}}</>){{end}}{{end}}{{end}}

  <info>{{ paddingName "help" }}</> Display help information

Use "<cyan>{$binName} {COMMAND} -h</>" for more information about a command
`
