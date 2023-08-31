package generics

import (
	"io"
	"sync/atomic"
)

// RefCloser is a reference counted io.Closer where the wrapped Closer is closed when its reference count reaches zero.
type RefCloser interface {

	// AddRef atomically increments the reference count.
	AddRef()

	// Release atomically decrements the reference count.
	// If the ref count is remains greater than zero, nil is returned.
	// If the ref count reaches zero, the underlying Closer was closed and it's error is returned.
	Close() error
}

// Wraps the given io.Closer into a RefCloser, initializing its reference count to 1.
func WrapInRefCloser(target io.Closer) RefCloser {
	rc := &refCloser{
		closer:   target,
		refCount: 1,
	}
	return rc
}

type refCloser struct {
	closer   io.Closer
	refCount int32
}

func (rc *refCloser) AddRef() {
	atomic.AddInt32(&rc.refCount, 1)
}

func (rc *refCloser) Close() error {
	if atomic.AddInt32(&rc.refCount, -1) > 0 {
		return nil
	}
	err := rc.closer.Close()
	rc.closer = nil
	return err
}
