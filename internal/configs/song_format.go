package configs

import "github.com/arcspace/go-librespot/Spotify"

type SongFormat string

const (
	Ogg_96  SongFormat = "OGG_96"
	Ogg_160 SongFormat = "OGG_160"
	Ogg_320 SongFormat = "OGG_320"
	Mp3_256 SongFormat = "MP3_256"
	Mp3_320 SongFormat = "MP3_320"
	Mp3_160 SongFormat = "MP3_160"
	Mp3_96  SongFormat = "MP3_96"
)

var formatMap = map[SongFormat]Spotify.AudioFile_Format{
	Ogg_96:  Spotify.AudioFile_OGG_VORBIS_96,
	Ogg_160: Spotify.AudioFile_OGG_VORBIS_160,
	Ogg_320: Spotify.AudioFile_OGG_VORBIS_320,
	Mp3_96:  Spotify.AudioFile_MP3_96,
	Mp3_160: Spotify.AudioFile_MP3_160,
	Mp3_320: Spotify.AudioFile_MP3_320,
	Mp3_256: Spotify.AudioFile_MP3_256,
}

func (f SongFormat) IsValid() bool {
	_, ok := formatMap[f]
	return ok
}

func (f SongFormat) ToSpotifyFormat() Spotify.AudioFile_Format {
	return formatMap[f]
}
