package lyrics

type LyricsResult struct {
	Lyrics *LyricsInfo `json:"lyrics"`
	Colors *ColorsInfo `json:"colors"`
}

type LyricsInfo struct {
	SyncType string       `json:"syncType"`
	Lines    []LyricsLine `json:"lines"`
}

type ColorsInfo struct {
	Background    int `json:"background"`
	Text          int `json:"text"`
	HighlightText int `json:"highlightText"`
}

type LyricsLine struct {
	Time  int    `json:"startTimeMs,string"`
	Words string `json:"words"`
}
