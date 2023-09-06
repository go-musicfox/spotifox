package configs

import "github.com/arcspace/go-librespot/Spotify"

type SongFormat string

const (
	Ogg_96  SongFormat = "OGG_96"
	Ogg_160 SongFormat = "OGG_160"
	Ogg_320 SongFormat = "OGG_320"
)

var formatMap = map[SongFormat]Spotify.AudioFile_Format{
	Ogg_96:  Spotify.AudioFile_OGG_VORBIS_96,
	Ogg_160: Spotify.AudioFile_OGG_VORBIS_160,
	Ogg_320: Spotify.AudioFile_OGG_VORBIS_320,
}

func (f SongFormat) IsValid() bool {
	_, ok := formatMap[f]
	return ok
}

func (f SongFormat) ToSpotifyFormat() Spotify.AudioFile_Format {
	return formatMap[f]
}
