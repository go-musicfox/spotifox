package Spotify

import (
	"fmt"
	"math/big"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const TrackUriPrefix = "spotify:track:"

var ErrInvalidURI = fmt.Errorf("failed to extract track ID from")

// ExtractAssetID returns the (ascii-based) hex track ID from a Spotify URI.
func ExtractAssetID(uri string) (trackID, trackHexID string, err error) {
	if strings.HasPrefix(uri, TrackUriPrefix) {
		trackID = uri[len(TrackUriPrefix):]
	} else {
		trackID = uri
	}

	if len(trackID) != 22 {
		err = fmt.Errorf("failed to extract track ID from %q", uri)
		return
	}

	trackHexID = fmt.Sprintf("%x", Convert62(trackID))
	return
}

func Convert62(id string) []byte {
	base := big.NewInt(62)

	n := &big.Int{}
	for _, c := range []byte(id) {
		d := big.NewInt(int64(strings.IndexByte(alphabet, c)))
		n = n.Mul(n, base)
		n = n.Add(n, d)
	}

	nBytes := n.Bytes()
	if len(nBytes) < 16 {
		paddingBytes := make([]byte, 16-len(nBytes))
		nBytes = append(paddingBytes, nBytes...)
	}
	return nBytes
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func ConvertTo62(raw []byte) string {
	bi := big.Int{}
	bi.SetBytes(raw)
	rem := big.NewInt(0)
	base := big.NewInt(62)
	zero := big.NewInt(0)
	result := ""

	for bi.Cmp(zero) > 0 {
		_, rem = bi.DivMod(&bi, base, rem)
		result += string(alphabet[int(rem.Uint64())])
	}

	for len(result) < 22 {
		result += "0"
	}
	return reverse(result)
}
