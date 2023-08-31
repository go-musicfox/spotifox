package connection

import (
	"encoding/binary"
	"io"
	"sync"
)

// PlainConnection represents an unencrypted connection to a Spotify AP
type PlainConnection struct {
	Writer io.Writer
	Reader io.Reader
	mutex  *sync.Mutex
}

// NewPlainConnection creates a new PlainConnection object.
func NewPlainConnection(reader io.Reader, writer io.Writer) PlainConnection {
	return PlainConnection{
		Reader: reader,
		Writer: writer,
		mutex:  &sync.Mutex{},
	}
}

func makePacketPrefix(prefix []byte, data []byte) []byte {
	size := len(prefix) + 4 + len(data)
	buf := make([]byte, 0, size)
	buf = append(buf, prefix...)
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(size))
	buf = append(buf, sizeBuf...)
	return append(buf, data...)
}

func (p *PlainConnection) SendPrefixPacket(prefix []byte, data []byte) ([]byte, error) {
	packet := makePacketPrefix(prefix, data)
	p.mutex.Lock()
	_, err := p.Writer.Write(packet)
	p.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	return packet, err
}

func (p *PlainConnection) RecvPacket() ([]byte, error) {
	var size uint32
	err := binary.Read(p.Reader, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	binary.BigEndian.PutUint32(buf, size)
	_, err = io.ReadFull(p.Reader, buf[4:])
	if err != nil {
		return nil, err
	}
	return buf, nil
}
