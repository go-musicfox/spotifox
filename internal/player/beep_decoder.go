package player

import (
	"io"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/vorbis"
	"github.com/pkg/errors"
)

func DecodeSong(t SongType, r io.ReadSeekCloser) (streamer beep.StreamSeekCloser, format beep.Format, err error) {
	switch t {
	case Mp3:
		streamer, format, err = mp3.Decode(r)
	case Ogg:
		streamer, format, err = vorbis.Decode(r)
	default:
		err = errors.Errorf("Unknown song type(%d)", t)
	}
	return
}
