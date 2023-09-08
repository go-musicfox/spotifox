package configs

type MainOptions struct {
	ShowTitle        bool
	LoadingText      string
	SongFormat       SongFormat
	PrimaryColor     string
	ShowLyric        bool
	LyricOffset      int
	ShowLyricTrans   bool
	ShowNotify       bool
	NotifyIcon       string
	PProfPort        int
	AltScreen        bool
	EnableMouseEvent bool
	DoubleColumn     bool
}
