package symbol

import (
	"sync/atomic"
)

func NewVolatileIssuer(startAt ID) Issuer {
	iss := &atomicIssuer{}
	iss.nextID.Store(uint32(startAt))
	iss.refCount.Store(1)
	return iss
}

// atomicIssuer implements symbol.Issuer using an atomic int
type atomicIssuer struct {
	refCount atomic.Int32
	nextID   atomic.Uint32
}

func (iss *atomicIssuer) IssueNextID() (ID, error) {
	if iss.refCount.Load() <= 0 {
		return 0, ErrIssuerNotOpen
	}
	nextID := iss.nextID.Add(1)
	return ID(nextID), nil
}

func (iss *atomicIssuer) AddRef() {
	if iss.refCount.Add(1) <= 1 {
		panic("AddRef() called on closed issuer")
	}
}

func (iss *atomicIssuer) Close() error {
	newRefCount := iss.refCount.Add(-1)
	if newRefCount < 0 {
		return ErrIssuerNotOpen
	}
	return nil
}
