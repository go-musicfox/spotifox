package symbol

import (
	"encoding/binary"
	"errors"

	"github.com/arcspace/go-arc-sdk/stdlib/generics"
)

// ID is a persistent integer value associated with an immutable string or buffer value.
// ID == 0 always maps to the empty string / buf.
type ID uint32

// Ord returns the ordinal value of this ID (a type recasting to uint32)
func (id ID) Ord() uint32 {
	return uint32(id)
}

// IDSz is the byte size of a symbol.ID (big endian)
// The tradeoff is between key bytes idle (wasted) in a massive db and exponentially more IDs available.
//
// The thinking of a 4 byte ID is that an symbol table exceeding 100 million entries is impractical and inefficient.
// If a billion symbol IDs is "not enough"  then you are issuing IDs for the wrong purpose.
const IDSz = 4

// DefaultIssuerMin specifies the default minimum ID value for newly issued IDs.
//
// ID values less than this value are reserved for clients to represent hard-wired or "out of band" meaning.
// "Hard-wired" meaning that Table.SetSymbolID() can be called with IDs less than MinIssuedID without risk
// of an auto-issued ID contending with it.
const DefaultIssuerMin = 600

type Issuer interface {
	generics.RefCloser

	// Issues the next sequential unique ID, starting at MinIssuedID.
	IssueNextID() (ID, error)
}

var ErrIssuerNotOpen = errors.New("issuer not open")

// Table stores value-ID pairs, designed for high-performance lookup of an ID or byte string.
// This implementation is intended to handle extreme loads, leveraging:
//   - ID-value pairs are cached once read, offering subsequent O(1) access
//   - Internal value allocations are pooled. The default TableOpts.PoolSz of 16k means
//     thousands of buffers can be issued or read under only a single allocation.
//
// All methods are thread-safe.
type Table interface {
	generics.RefCloser

	// Returns the Issuer being used by this Table (passed via TableOpts.Issuer or auto-created if no TableOpts.Issuer was given)
	// Note that retained references should make use of generics.RefCloser to ensure proper closure.
	Issuer() Issuer

	// Returns the symbol ID previously associated with the given string/buffer value.
	// The given value buffer is never retained.
	//
	// If not found and autoIssue == true, a new entry is created and the new ID returned.
	// Newly issued IDs are always > 0 and use the lower bytes of the returned ID (see type ID comments).
	//
	// If not found and autoIssue == false, 0 is returned.
	GetSymbolID(value []byte, autoIssue bool) ID

	// Associates the given buffer value to the given symbol ID, allowing multiple values to be mapped to a single ID.
	// If ID == 0, then this is the equivalent to GetSymbolID(value, true).
	SetSymbolID(value []byte, ID ID) ID

	// Looks up and appends the byte string associated with the given symbol ID to the given buf.
	// If ID is invalid or not found, nil is returned.
	GetSymbol(ID ID, io []byte) []byte
}

// Reads a big endian encoded uint32 ID from the given byte slice
func ReadID(in []byte) (uint32, []byte) {
	ID := binary.BigEndian.Uint32(in)
	return ID, in[IDSz:]
}

// Reads an ID from the given byte slice (reading IDSz=4 bytes)
func (id *ID) ReadFrom(in []byte) {
	*id = ID(binary.BigEndian.Uint32(in))
}

func AppendID(io []byte, ID uint32) []byte {
	return append(io, // big endian marshal
		byte(ID>>24),
		byte(ID>>16),
		byte(ID>>8),
		byte(ID))
}

func (id ID) AppendTo(io []byte) []byte {
	return append(io, // big endian marshal
		byte(uint32(id)>>24),
		byte(uint32(id)>>16),
		byte(uint32(id)>>8),
		byte(id))
}
