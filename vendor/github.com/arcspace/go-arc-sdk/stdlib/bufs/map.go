package bufs

import (
	"bytes"
	"hash/maphash"
)

type MapEntry interface {
	MapKey() []byte
}

type BufMap struct {
	hashMap map[uint64]MapEntry
	hasher  maphash.Hash
}

func NewBufMap() BufMap {
	return BufMap{
		hashMap: make(map[uint64]MapEntry),
	}
}

// Returns the entry having a matching key (or nil)
func (m *BufMap) Get(key []byte) MapEntry {
	existing, _ := m.get(key)
	return existing
}

// Returns the entry being replaced (or nil)
func (m *BufMap) Put(entry MapEntry) MapEntry {
	existing, atHash := m.get(entry.MapKey())

	m.hashMap[atHash] = entry
	return existing
}

// Removes and returns the entry having a matching key (or nil)
func (m *BufMap) Remove(key []byte) MapEntry {
	existing, atHash := m.get(key)

	if existing != nil {
		delete(m.hashMap, atHash)
	}
	return existing
}

func (m *BufMap) get(key []byte) (entry MapEntry, atHash uint64) {
	m.hasher.Reset()
	m.hasher.Write(key)
	atHash = m.hasher.Sum64()

	var found bool
	entry, found = m.hashMap[atHash]
	for found {
		if bytes.Equal(entry.MapKey(), key) {
			return
		}
		atHash++
		entry, found = m.hashMap[atHash]
	}

	return nil, atHash
}
