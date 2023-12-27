module github.com/go-musicfox/spotifox

go 1.21

require (
	github.com/anhoder/foxful-cli v0.1.4
	github.com/arcspace/go-arc-sdk v0.0.0-20230811172934-db6c05cc94b2
	github.com/arcspace/go-librespot v0.0.0-20230811173922-2e901b172fbe
	github.com/buger/jsonparser v1.1.1
	github.com/charmbracelet/bubbles v0.16.1
	github.com/charmbracelet/bubbletea v0.24.2
	github.com/charmbracelet/lipgloss v0.8.0
	github.com/ebitengine/purego v0.4.0
	github.com/faiface/beep v1.1.0
	github.com/go-musicfox/notificator v0.1.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/gookit/gcli/v2 v2.3.4
	github.com/gookit/ini/v2 v2.2.2
	github.com/mattn/go-runewidth v0.0.15
	github.com/muesli/termenv v0.15.2
	github.com/nicksnyder/go-i18n/v2 v2.2.1
	github.com/pkg/errors v0.9.1
	github.com/raitonoberu/lyricsapi v0.0.0-20230113141433-eded40b42d7c
	github.com/shkh/lastfm-go v0.0.0-20191215035245-89a801c244e0
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966
	github.com/zmb3/spotify/v2 v2.3.1
	go.etcd.io/bbolt v1.3.7
	golang.org/x/oauth2 v0.7.0
)

require (
	capnproto.org/go/capnp/v3 v3.0.0-alpha-29 // indirect
	github.com/brynbellomy/klog v0.0.0-20200414031930-87fbf2e555ae // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/rs/cors v1.9.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.57.1 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	zenhack.net/go/util v0.0.0-20230607025951-8b02fee814ae // indirect
)

require (
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/containerd/console v1.0.4-0.20230313162750-1ae8d489ac81 // indirect
	github.com/fogleman/ease v0.0.0-20170301025033-8da417bf1776 // indirect
	github.com/gookit/color v1.5.3 // indirect
	github.com/gookit/goutil v0.6.10 // indirect
	github.com/hajimehoshi/go-mp3 v0.3.4 // indirect
	github.com/hajimehoshi/oto v1.0.1 // indirect
	github.com/jfreymuth/oggvorbis v1.0.5 // indirect
	github.com/jfreymuth/vorbis v1.0.2 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sahilm/fuzzy v0.1.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df // indirect
	golang.org/x/exp/shiny v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/image v0.8.0 // indirect
	golang.org/x/mobile v0.0.0-20230531173138-3c911d8e3eda // indirect
	golang.org/x/mod v0.12.0
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0
)

replace (
	github.com/arcspace/go-librespot v0.0.0-20230811173922-2e901b172fbe => github.com/go-musicfox/go-librespot v0.1.0
	github.com/charmbracelet/bubbletea v0.24.2 => github.com/go-musicfox/bubbletea v0.24.1
	github.com/cnsilvan/UnblockNeteaseMusic v0.0.0-20230310083816-92b59c95a366 => github.com/go-musicfox/UnblockNeteaseMusic v0.1.2
	github.com/faiface/beep v1.1.0 => github.com/go-musicfox/beep v1.2.4
	github.com/frolovo22/tag v0.0.2 => github.com/go-musicfox/tag v1.0.2
	github.com/gookit/gcli/v2 v2.3.4 => github.com/anhoder/gcli/v2 v2.3.5
	github.com/hajimehoshi/go-mp3 v0.3.4 => github.com/go-musicfox/go-mp3 v0.3.3
	github.com/hajimehoshi/oto v1.0.1 => github.com/go-musicfox/oto v1.0.3
	github.com/shkh/lastfm-go => github.com/go-musicfox/lastfm-go v0.0.2
)
