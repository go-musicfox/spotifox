package asset

import (
	"bytes"
	"crypto/aes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/arcspace/go-arc-sdk/apis/arc"
	"github.com/arcspace/go-arc-sdk/stdlib/errors"
	"github.com/arcspace/go-arc-sdk/stdlib/task"
	"github.com/arcspace/go-librespot/Spotify"
	"github.com/arcspace/go-librespot/librespot/core/connection"
	"github.com/arcspace/go-librespot/librespot/mercury"
)

type Downloader interface {
	HandleCmd(cmd byte, data []byte) error

	// Blocks until the asset is ready to be accessed.
	PinTrack(uri string) (arc.MediaAsset, error)

	SetAudioFormat(f Spotify.AudioFile_Format)
}

type downloader struct {
	task.Context
	stream  connection.PacketStream
	mercury *mercury.Client

	chMu        sync.Mutex
	chMap       map[uint16]*assetChunk
	seqChans    sync.Map // TODO: make this less gross
	nextChan    uint16
	audioFormat Spotify.AudioFile_Format
}

var extMap = map[Spotify.AudioFile_Format]string{
	Spotify.AudioFile_OGG_VORBIS_96:  ".96.ogg",
	Spotify.AudioFile_OGG_VORBIS_160: ".160.ogg",
	Spotify.AudioFile_OGG_VORBIS_320: ".320.ogg",
	Spotify.AudioFile_MP3_256:        ".256.mp3",
	Spotify.AudioFile_MP3_320:        ".320.mp3",
	Spotify.AudioFile_MP3_160:        ".160.mp3",
	Spotify.AudioFile_MP3_96:         ".96.mp3",
	Spotify.AudioFile_MP3_160_ENC:    ".160enc.mp3",
	Spotify.AudioFile_AAC_24:         ".24.aac",
	Spotify.AudioFile_AAC_48:         ".48.aac",
}

func NewDownloader(conn connection.PacketStream, client *mercury.Client) Downloader {
	dl := &downloader{
		stream:      conn,
		mercury:     client,
		chMap:       map[uint16]*assetChunk{},
		seqChans:    sync.Map{},
		chMu:        sync.Mutex{},
		nextChan:    0,
		audioFormat: Spotify.AudioFile_OGG_VORBIS_160,
	}
	return dl
}

func (dl *downloader) SetAudioFormat(f Spotify.AudioFile_Format) {
	dl.audioFormat = f
}

func (dl *downloader) PinTrack(assetURI string) (arc.MediaAsset, error) {
	// Get the track metadata: it holds information about which files and encodings are available
	assetID, track, err := dl.mercury.GetTrack(assetURI)
	if err != nil {
		return nil, errors.Wrap(err, "error getting track")
	}

	// As of May 2023, fetching AAC 160 or 320 result in a PacketAesKeyError for reasons unknown.
	// Stranger still, AAC_48 returns a key but the data appears to be corrupt.
	// Posts such as https://github.com/librespot-org/librespot-golang/issues/28 suggest it has been broken for a while.
	// Unknown: does AAC work on https://github.com/librespot-org/librespot
	trackFile := dl.chooseBestFile(track)
	if trackFile == nil {
		err = fmt.Errorf("no file found for format %v", dl.audioFormat)
		return nil, err
	}

	asset := newMediaAsset(dl, track)

	asset.trackFile = trackFile
	ext := extMap[trackFile.GetFormat()]
	if ext == "" {
		return nil, fmt.Errorf("unknown format: %d", trackFile.GetFormat())
	}

	switch {
	case strings.HasSuffix(ext, ".mp3"):
		asset.mediaType = "audio/mpeg"
	case strings.HasSuffix(ext, ".ogg"):
		asset.mediaType = "audio/ogg"
		asset.assetByteOfs = SPOTIFY_OGG_HEADER_SIZE
	case strings.HasSuffix(ext, ".aac"):
		asset.mediaType = "audio/aac"
	}

	asset.label = fmt.Sprintf("%s - %s (%s)%s", asset.track.GetArtist()[0].GetName(), asset.track.GetName(), assetID, ext)

	err = dl.loadTrackKey(asset)
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %+v", err)
	}

	return asset, err
}

func (dl *downloader) chooseBestFile(track *Spotify.Track) *Spotify.AudioFile {
	for _, file := range track.File {
		if file.GetFormat() == dl.audioFormat {
			return file
		}
	}
	for _, alt := range track.Alternative {
		for _, file := range alt.File {
			if file.GetFormat() == dl.audioFormat {
				return file
			}
		}
	}
	return nil
}

func (dl *downloader) loadTrackKey(asset *mediaAsset) error {
	seqInt, _ := dl.mercury.NextSeqWithInt()

	channel := make(chan []byte)
	dl.seqChans.Store(seqInt, channel)

	req := buildKeyRequest(seqInt, asset.track.Gid, asset.trackFile.FileId)
	err := dl.stream.SendPacket(connection.PacketRequestKey, req)
	if err != nil {
		log.Println("error sending packet", err)
		return err
	}

	key := <-channel
	dl.seqChans.Delete(seqInt)

	asset.cipher, err = aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to decrypt aes cipher: %+v", err)
	}

	return nil
}

// opens a new data channel to recv the requested chunk
func (dl *downloader) RequestChunk(chunkIdx ChunkIdx, asset *mediaAsset) (*assetChunk, error) {

	dl.chMu.Lock()
	chunk := &assetChunk{
		chID:   dl.nextChan,
		asset:  asset,
		Idx:    chunkIdx,
		status: chunkInProgress,
	}
	dl.nextChan++
	dl.chMap[chunk.chID] = chunk
	dl.chMu.Unlock()

	if cap(chunk.Data) < kChunkByteSize {
		chunk.Data = make([]byte, 0, kChunkByteSize)
	} else {
		chunk.Data = chunk.Data[:0]
	}

	wordOfs := uint32(chunkIdx * kChunkWordSize)

	if err := dl.stream.SendPacket(
		connection.PacketStreamChunk,
		buildAudioChunkRequest(
			chunk.chID,
			asset.trackFile.FileId,
			wordOfs,
			wordOfs+kChunkWordSize,
		),
	); err != nil {
		return nil, fmt.Errorf("could not send stream chunk: %+v", err)
	}

	return chunk, nil
}

func (dl *downloader) HandleCmd(cmd byte, data []byte) error {
	switch {
	case cmd == connection.PacketAesKey:
		seqID := binary.BigEndian.Uint32(data[0:4])
		if channel, ok := dl.seqChans.Load(seqID); ok {
			channel.(chan []byte) <- data[4:20]
		} else {
			return fmt.Errorf("unknown channel %d", seqID)
		}
	case cmd == connection.PacketAesKeyError:
		seqID := binary.BigEndian.Uint32(data[0:4])
		if channel, ok := dl.seqChans.Load(seqID); ok {
			channel.(chan []byte) <- nil
		} else {
			return fmt.Errorf("unknown channel %d", seqID)
		}
		return fmt.Errorf("audio key error")
	case cmd == connection.PacketStreamChunkRes:
		chID, assetData, err := ReadU16(data)
		if err != nil {
			return err
		}
		dl.chMu.Lock()
		dstCh, ok := dl.chMap[chID]
		dl.chMu.Unlock()
		if ok {
			dl.handlePacket(dstCh, assetData)
		} else {
			return fmt.Errorf("unknown channel")
		}
	}
	return nil
}

func ReadU32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func ReadU16(data []byte) (uint16, []byte, error) {
	if len(data) < 2 {
		return 0, nil, fmt.Errorf("not enough data")
	}
	return binary.BigEndian.Uint16(data), data[2:], nil
}

func (dl *downloader) releaseChannel(chID uint16) {
	dl.chMu.Lock()
	delete(dl.chMap, chID)
	dl.chMu.Unlock()
}

const (
	Header_AssetWordSize uint8 = 0x03
)

func (dl *downloader) handlePacket(chunk *assetChunk, data []byte) {

	if !chunk.gotHeader {
		reader := bytes.NewReader(data)
		//bytesRead := uint16(0)

		for {
			sectLen := uint16(0)
			err := binary.Read(reader, binary.BigEndian, &sectLen)
			if sectLen == 0 || err != nil {
				break
			}

			ofs, _ := reader.Seek(0, io.SeekCurrent)
			if err != nil {
				break
			}

			sectTypeID := data[ofs]
			ofs++

			switch sectTypeID {
			case Header_AssetWordSize:
				wordSz := binary.BigEndian.Uint32(data[ofs:])
				chunk.totalAssetSz = int64(wordSz) << 2
			}

			reader.Seek(int64(sectLen), io.SeekCurrent)
		}

		chunk.gotHeader = true
	} else {

		// is there a more robust way to signal completion?
		if len(data) == 0 {
			chunk.status = chunkReadyToDecrypt
			select {
			case chunk.asset.onChunkComplete <- chunk:
			case <-chunk.asset.Closing():
			}
			dl.releaseChannel(chunk.chID)
		} else {
			chunk.Data = append(chunk.Data, data...)
		}

	}

}
