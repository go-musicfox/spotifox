package bufs

import (
	"encoding/base32"
	"encoding/hex"
	"encoding/json"

	//"github.com/mmcloughlin/geohash"

	"reflect"
)

// GeohashBase32Alphabet is the standard geo-hash alphabet used for Base32Encoding.
// It chooses particular characters that are not visually similar to each other.
const GeohashBase32Alphabet = "0123456789bcdefghjkmnpqrstuvwxyz"

var (
	// Base32Encoding is used to encode/decode binary buffer to/from base 32
	Base32Encoding = base32.NewEncoding(GeohashBase32Alphabet).WithPadding(base32.NoPadding)

	// GenesisMemberID is the genesis member ID
	GenesisMemberID = uint32(1)
)

// Zero zeros out a given slice
func Zero(buf []byte) {
	N := int32(len(buf))
	for i := int32(0); i < N; i++ {
		buf[i] = 0
	}
}

// Marshaler generalizes efficient serialization
type Marshaler interface {
	Marshal() ([]byte, error)
	MarshalToSizedBuffer([]byte) (int, error)
	Size() int
}

// Unmarshaler used to generalize deserialization
type Unmarshaler interface {
	Unmarshal([]byte) error
}

// SmartMarshal marshals the given item to the given buffer.  If there is not enough space a new one is allocated.  The purpose of this is to reuse a scrap buffer.
func SmartMarshal(item Marshaler, tryDst []byte) []byte {
	bufSz := cap(tryDst)
	encSz := item.Size()
	if encSz > bufSz {
		bufSz = (encSz + 15) &^ 15
		tryDst = make([]byte, bufSz)
	}

	var err error
	encSz, err = item.MarshalToSizedBuffer(tryDst[:encSz])
	if err != nil {
		panic(err)
	}

	return tryDst[:encSz]
}

// SmartMarshalToBase32 marshals the given item and then encodes it into a base32 (ASCII) byte string.
//
// If tryDst is not large enough, a new buffer is allocated and returned in its place.
func SmartMarshalToBase32(item Marshaler, tryDst []byte) []byte {
	bufSz := cap(tryDst)
	binSz := item.Size()
	{
		safeSz := 4 + 4*((binSz+2)/3)
		if safeSz > bufSz {
			bufSz = (safeSz + 7) &^ 7
			tryDst = make([]byte, bufSz)
		}
	}

	// First, marshal the item to the right-side of the scrap buffer
	binBuf := tryDst[bufSz-binSz : bufSz]
	var err error
	binSz, err = item.MarshalToSizedBuffer(binBuf)
	if err != nil {
		panic(err)
	}

	// Now encode the marshaled to the left side of the scrap buffer.
	// There is overlap, but encoding consumes from left to right, so it's safe.
	encSz := Base32Encoding.EncodedLen(binSz)
	tryDst = tryDst[:encSz]
	Base32Encoding.Encode(tryDst, binBuf[:binSz])

	return tryDst
}

// SmartDecodeFromBase32 decodes the base32 (ASCII) string into the given scrap buffer, returning the scrap buffer set to proper size.
//
// If tryDst is not large enough, a new buffer is allocated and returned in its place.
func SmartDecodeFromBase32(srcBase32 []byte, tryDst []byte) ([]byte, error) {
	binSz := Base32Encoding.DecodedLen(len(srcBase32))

	bufSz := cap(tryDst)
	if binSz > bufSz {
		bufSz = (binSz + 7) &^ 7
		tryDst = make([]byte, bufSz)
	}
	var err error
	binSz, err = Base32Encoding.Decode(tryDst[:binSz], srcBase32)
	return tryDst[:binSz], err
}

// Buf is a flexible buffer designed for reuse.
type Buf struct {
	Unmarshaler

	Bytes []byte
}

// Unmarshal effectively copies the src buffer.
func (buf *Buf) Unmarshal(srcBuf []byte) error {
	N := len(srcBuf)
	if cap(buf.Bytes) < N {
		allocSz := ((N + 127) >> 7) << 7
		buf.Bytes = make([]byte, N, allocSz)
	} else {
		buf.Bytes = buf.Bytes[:N]
	}
	copy(buf.Bytes, srcBuf)

	return nil
}

var (
	bytesType = reflect.TypeOf(Bytes(nil))
)

// Bytes marshal/unmarshal as a JSON string with 0x prefix.
// The empty slice marshals as "0x".
type Bytes []byte

// MarshalText implements encoding.TextMarshaler
func (b Bytes) MarshalText() ([]byte, error) {
	out := make([]byte, len(b)*2+2)
	out[0] = '0'
	out[1] = 'x'
	hex.Encode(out[2:], b)
	return out, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bytes) UnmarshalJSON(in []byte) error {
	if !isString(in) {
		return errNonString(bytesType)
	}
	return wrapTypeError(b.UnmarshalText(in[1:len(in)-1]), bytesType)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Bytes) UnmarshalText(input []byte) error {
	raw, err := checkText(input)
	if err != nil {
		return err
	}
	dec := make([]byte, len(raw)/2)
	if _, err = hex.Decode(dec, raw); err == nil {
		*b = dec
	}
	return err
}

// String returns the hex encoding of b.
func (b Bytes) String() string {
	out := make([]byte, len(b)*2+2)
	out[0] = '0'
	out[1] = 'x'
	hex.Encode(out[2:], b)
	return string(out)
}

type encodingErr struct {
	msg string
}

func (err *encodingErr) Error() string {
	return err.msg
}

// Errors
var (
	ErrSyntax = &encodingErr{"invalid hex string"}
)

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

func wrapTypeError(err error, typ reflect.Type) error {
	if _, ok := err.(*encodingErr); ok {
		return &json.UnmarshalTypeError{Value: err.Error(), Type: typ}
	}
	return err
}

func errNonString(typ reflect.Type) error {
	return &json.UnmarshalTypeError{Value: "non-string", Type: typ}
}

func checkText(in []byte) ([]byte, error) {
	N := len(in)
	if N == 0 {
		return nil, nil // empty strings are allowed
	}
	if N >= 2 && in[0] == '0' && (in[1] == 'x' || in[1] == 'X') {
		in = in[2:]
		N -= 2
	}
	return in, nil
}

// BufDesc returns a base32 encoding of a binary string, limiting it to a short number of character for debugging and logging.
func BufDesc(inBuf []byte) string {
	if len(inBuf) == 0 {
		return "nil"
	}

	buf := inBuf

	const limit = 12
	alreadyASCII := true
	for _, b := range buf {
		if b < 32 || b > 126 {
			alreadyASCII = false
			break
		}
	}

	suffix := ""
	if len(buf) > limit {
		buf = buf[:limit]
		suffix = "…"
	}

	outStr := ""
	if alreadyASCII {
		outStr = string(buf)
	} else {
		outStr = Base32Encoding.EncodeToString(buf)
	}

	return outStr + suffix
}
