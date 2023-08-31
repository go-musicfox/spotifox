package arc

import (
	"bytes"
	strings "strings"
	"time"

	"github.com/arcspace/go-arc-sdk/stdlib/bufs"
)

// TimeID is a locally unique UTC16 value -- see SessionRegistry.IssueTimeID()
type TimeID UTC16

// CellID is a uniquely issued TimeID (guaranteed to be globally unique) used to persistently identify a Cell.
type CellID TimeID

type CellTID struct {
	TID_UTC16 TimeID
	TID_HASH1 uint64
	TID_HASH2 uint64
	TID_HASH3 uint64
}

// TID identifies a specific planet, node, or transaction.
//
// Unless otherwise specified a TID in the wild should always be considered read-only.
type TID []byte

// TxID is embedded UTC16 value followed by a 24 byte hash.
type TxID [TIDBinaryLen]byte

// Byte size of a TID, a hash with a leading embedded big endian binary time index.
const TIDBinaryLen = int(Const_TIDBinaryLen)

// ASCII-compatible string length of a (binary) TID encoded into its base32 form.
const TIDStringLen = int(Const_TIDStringLen)

// nilTID is a zeroed TID that denotes a void/nil/zero value of a TID
var nilTID = TxID{}

// UTC16 is a signed UTC timestamp, storing the elapsed 1/65536 second ticks since Jan 1, 1970 UTC.
//
// Shifting this value to the right 16 bits will yield standard Unix time.
// This means there are 47 bits dedicated for seconds, implying a max timestamp of 4.4 million years.
type UTC16 int64

const (
	SI_DistantFuture = UTC16(0x7FFFFFFFFFFFFFFF)
)

// Converts a time.Time to a UTC16.
func ConvertToUTC(t time.Time) UTC16 {
	time16 := t.Unix() << 16
	frac := uint16((2199 * (uint32(t.Nanosecond()) >> 10)) >> 15)
	return UTC16(time16 | int64(frac))
}

// Converts milliseconds to UTC16.
func ConvertMsToUTC(ms int64) UTC16 {
	return UTC16((ms << 16) / 1000)
}

// Converts UTC16 to a time.Time.
func (t UTC16) ToTime() time.Time {
	return time.Unix(int64(t>>16), int64(t&0xFFFF)*15259)
}

// Converts UTC16 to milliseconds.
func (t UTC16) ToMs() int64 {
	return (int64(t>>8) * 1000) >> 8
}

// TID is a convenience function that returns the TID contained within this TxID.
func (tid *TxID) TID() TID {
	return tid[:]
}

// Base32 returns this TID in Base32 form.
func (tid *TxID) Base32() string {
	return bufs.Base32Encoding.EncodeToString(tid[:])
}

// IsNil returns true if this TID length is 0 or is equal to NilTID
func (tid TID) IsNil() bool {
	if len(tid) == 0 {
		return true
	}

	if bytes.Equal(tid, nilTID[:]) {
		return true
	}

	return false
}

// Clone returns a duplicate of this TID
func (tid TID) Clone() TID {
	dupe := make([]byte, len(tid))
	copy(dupe, tid)
	return dupe
}

// Buf is a convenience function that make a new TxID from a TID byte slice.
func (tid TID) Buf() TxID {
	var blob TxID
	copy(blob[:], tid)
	return blob
}

// Base32 returns this TID in Base32 form.
func (tid TID) Base32() string {
	return bufs.Base32Encoding.EncodeToString(tid)
}

// Appends the base 32 ASCII encoding of this TID to the given buffer
func (tid TID) AppendAsBase32(in []byte) []byte {
	encLen := bufs.Base32Encoding.EncodedLen(len(tid))
	needed := len(in) + encLen
	buf := in
	if needed > cap(buf) {
		buf = make([]byte, (needed+0x100)&^0xFF)
		buf = append(buf[:0], in...)
	}
	buf = buf[:needed]
	bufs.Base32Encoding.Encode(buf[len(in):needed], tid)
	return buf
}

// SuffixStr returns the last few digits of this TID in string form (for easy reading, logs, etc)
func (tid TID) SuffixStr() string {
	const summaryStrLen = 5

	R := len(tid)
	L := R - summaryStrLen
	if L < 0 {
		L = 0
	}
	return bufs.Base32Encoding.EncodeToString(tid[L:R])
}

// SetTimeAndHash writes the given timestamp and the right-most part of inSig into this TID.
//
// See comments for TIDBinaryLen
func (tid TID) SetTimeAndHash(time UTC16, hash []byte) {
	tid.SetUTC(time)
	tid.SetHash(hash)
}

// SetHash sets the sig/hash portion of this ID
func (tid TID) SetHash(hash []byte) {
	const TIDHashSz = int(Const_TIDBinaryLen - Const_TIDTimestampSz)
	pos := len(hash) - TIDHashSz
	if pos >= 0 {
		copy(tid[TIDHashSz:], hash[pos:])
	} else {
		for i := 8; i < int(Const_TIDBinaryLen); i++ {
			tid[i] = hash[i]
		}
	}
}

// SetUTC16 writes the given UTC16 into this TID
func (tid TID) SetUTC(t UTC16) {
	tid[0] = byte(t >> 56)
	tid[1] = byte(t >> 48)
	tid[2] = byte(t >> 40)
	tid[3] = byte(t >> 32)
	tid[4] = byte(t >> 24)
	tid[5] = byte(t >> 16)
	tid[6] = byte(t >> 8)
	tid[7] = byte(t)
}

// ExtractUTC16 returns the unix timestamp embedded in this TID (a unix timestamp in 1<<16 seconds UTC)
func (tid TID) ExtractUTC() UTC16 {
	t := int64(tid[0])
	t = (t << 8) | int64(tid[1])
	t = (t << 8) | int64(tid[2])
	t = (t << 8) | int64(tid[3])
	t = (t << 8) | int64(tid[4])
	t = (t << 8) | int64(tid[5])
	t = (t << 8) | int64(tid[6])
	t = (t << 8) | int64(tid[7])

	return UTC16(t)
}

// ExtractTime returns the unix timestamp embedded in this TID (a unix timestamp in seconds UTC)
func (tid TID) ExtractTime() int64 {
	t := int64(tid[0])
	t = (t << 8) | int64(tid[1])
	t = (t << 8) | int64(tid[2])
	t = (t << 8) | int64(tid[3])
	t = (t << 8) | int64(tid[4])
	t = (t << 8) | int64(tid[5])

	return t
}

// SelectEarlier looks in inTime a chooses whichever is earlier.
//
// If t is later than the time embedded in this TID, then this function has no effect and returns false.
//
// If t is earlier, then this TID is initialized to t (and the rest zeroed out) and returns true.
func (tid TID) SelectEarlier(t UTC16) bool {

	TIDt := tid.ExtractUTC()

	// Timestamp of 0 is reserved and should only reflect an invalid/uninitialized TID.
	if t < 0 {
		t = 0
	}

	if t < TIDt || t == 0 {
		tid.SetUTC(t)
		for i := 8; i < len(tid); i++ {
			tid[i] = 0
		}
		return true
	}

	return false
}

// CopyNext copies the given TID and increments it by 1, typically useful for seeking the next entry after a given one.
func (tid TID) CopyNext(inTID TID) {
	copy(tid, inTID)
	for j := len(tid) - 1; j > 0; j-- {
		tid[j]++
		if tid[j] > 0 {
			break
		}
	}
}

func (id ConstSymbol) Ord() uint32 {
	return uint32(id)
}

// Analyses an AttrSpec's SeriesSpec and returns the index class it uses.
func GetSeriesIndexType(seriesSpec string) SeriesIndexType {
	switch {
	case strings.HasSuffix(seriesSpec, ".Name"):
		return SeriesIndexType_Name
	default:
		return SeriesIndexType_Literal
	}
}

func (params *PinReqParams) URLPath() []string {
	if params.URL == nil {
		return nil
	}
	path := params.URL.Path
	if path != "" && path[0] == '/' {
		path = path[1:]
	}
	return strings.Split(path, "/")
}

func (params *PinReqParams) Params() *PinReqParams {
	return params
}

/*

// ReadCell loads a cell with the given URI having the inferred schema (built from its fields using reflection).
// The URI is scoped into the user's home planet and AppID.
func ReadCell(ctx AppContext, subKey string, schema *AttrSchema, dstStruct any) error {

	dst := reflect.Indirect(reflect.ValueOf(dstStruct))
	switch dst.Kind() {
	case reflect.Pointer:
		dst = dst.Elem()
	case reflect.Struct:
	default:
		return ErrCode_ExportErr.Errorf("expected struct, got %v", dst.Kind())
	}

	var keyBuf [128]byte
	cellKey := append(append(keyBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

	msgs := make([]*Msg, 0, len(schema.Attrs))
	err := ctx.User().HomePlanet().ReadCell(cellKey, schema, func(msg *Msg) {
		switch msg.Op {
		case MsgOp_PushAttr:
			msgs = append(msgs, msg)
		}
	})
	if err != nil {
		return err
	}

	numFields := dst.NumField()
	valType := dst.Type()

	for fi := 0; fi < numFields; fi++ {
		field := valType.Field(fi)
		for _, ai := range schema.Attrs {
			if ai.TypedName == field.Name {
				for _, msg := range msgs {
					if msg.AttrID == ai.AttrID {
						msg.LoadVal(dst.Field(fi).Addr().Interface())
						goto nextField
					}
				}
			}
		}
	nextField:
	}
	return err
}

// WriteCell is the write analog of ReadCell.
func WriteCell(ctx AppContext, subKey string, schema *AttrSchema, srcStruct any) error {

	src := reflect.Indirect(reflect.ValueOf(srcStruct))
	switch src.Kind() {
	case reflect.Pointer:
		src = src.Elem()
	case reflect.Struct:
	default:
		return ErrCode_ExportErr.Errorf("expected struct, got %v", src.Kind())
	}

	{
		tx := NewMsgBatch()
		msg := tx.AddMsg()
		msg.Op = MsgOp_UpsertCell
		msg.ValType = ValType_SchemaID.Ord()
		msg.ValInt = int64(schema.SchemaID)
		msg.ValBuf = append(append(msg.ValBuf[:0], []byte(ctx.StateScope())...), []byte(subKey)...)

		numFields := src.NumField()
		valType := src.Type()

		for _, attr := range schema.Attrs {
			msg := tx.AddMsg()
			msg.Op = MsgOp_PushAttr
			msg.AttrID = attr.AttrID
			for i := 0; i < numFields; i++ {
				if valType.Field(i).Name == attr.TypedName {
					msg.setVal(src.Field(i).Interface())
					break
				}
			}
			if msg.ValType == ValType_nil.Ord() {
				panic("missing field")
			}
		}

		msg = tx.AddMsg()
		msg.Op = MsgOp_Commit

		if err := ctx.User().HomePlanet().PushTx(tx); err != nil {
			return err
		}
	}

	return nil
}


func (req *CellReq) GetKwArg(argKey string) (string, bool) {
	for _, arg := range req.Args {
		if arg.Key == argKey {
			if arg.Val != "" {
				return arg.Val, true
			}
			return string(arg.ValBuf), true
		}
	}
	return "", false
}

func (req *CellReq) GetChildSchema(modelURI string) *AttrSchema {
	for _, schema := range req.ChildSchemas {
		if schema.CellDataModel == modelURI {
			return schema
		}
	}
	return nil
}

func (req *CellReq) PushBeginPin(target CellID) {
	m := NewMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PinCell
	req.PushUpdate(m)
}

func (req *CellReq) PushInsertCell(target CellID, schema *AttrSchema) {
	if schema != nil {
		m := NewMsg()
		m.CellID = target.U64()
		m.Op = MsgOp_InsertChildCell
		m.ValType = int32(ValType_SchemaID)
		m.ValInt = int64(schema.SchemaID)
		req.PushUpdate(m)
	}
}

// Pushes the given attr to the client
func (req *CellReq) PushAttr(target CellID, schema *AttrSchema, attrURI string, val Value) {
	attr := schema.LookupAttr(attrURI)
	if attr == nil {
		return
	}

	m := NewMsg()
	m.CellID = target.U64()
	m.Op = MsgOp_PushAttr
	m.AttrID = attr.AttrID
	if attr.SeriesType == SeriesType_Fixed {
		m.SI = attr.BoundSI
	}
	val.MarshalToMsg(m)
	if attr.ValTypeID != 0 { // what is this for!?
		m.ValType = int32(attr.ValTypeID)
	}
	req.PushUpdate(m)
}

func (req *CellReq) PushCheckpoint(err error) {
	m := NewMsg()
	m.Op = MsgOp_Commit
	m.CellID = req.PinCell.U64()
	if err != nil {
		m.setVal(err)
	}
	req.PushUpdate(m)
}

*/
