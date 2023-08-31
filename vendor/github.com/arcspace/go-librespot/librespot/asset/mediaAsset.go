package asset

import (
	"crypto/cipher"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/arcspace/go-arc-sdk/apis/arc"
	"github.com/arcspace/go-arc-sdk/stdlib/task"
	"github.com/arcspace/go-librespot/Spotify"
	"github.com/arcspace/go-librespot/librespot/core/crypto"
)

type ChunkIdx int32 // 0-based index of the chunk in the asset

const (
	kChunkWordSize = 1 << 15             // Number of 4-byte words per chunk
	kChunkByteSize = kChunkWordSize << 2 // 4 bytes per word

	// Spotify inserts a custom Ogg packet at the start with custom metadata values, that you would
	// otherwise expect in Vorbis comments. This packet isn't well-formed and players may balk at it.
	// Search for "parse_from_ogg" in librespot -- https://github.com/librespot-org/librespot
	SPOTIFY_OGG_HEADER_SIZE = 0xa7
)

type ChunkStatus int32

const (
	chunkHalted ChunkStatus = iota
	chunkInProgress
	chunkReadyToDecrypt // chunk obtained but is still encrypted
	chunkReady          // chunk obtained and decrypted successfully (chunkData is ready and the correct size)

	ChunkIdx_Nil          = ChunkIdx(-1)
	kMaxConcurrentFetches = 1
)

const kDebug = false

type assetChunk struct {
	chID         uint16
	totalAssetSz int64
	gotHeader    bool
	asset        *mediaAsset
	Data         []byte
	Idx          ChunkIdx
	status       ChunkStatus
	accessStamp  int64        // allows stale chunks to be identified and evicted
	numReaders   atomic.Int32 // tracks what it in use

}

// mediaAsset represents a downloadable/cached audio file fetched by Spotify, in an encoded format (OGG, etc)
type mediaAsset struct {
	task.Context

	label        string
	mediaType    string
	assetByteSz  int64 // contiguous chunk byte span; 0 denotes chunk #0 is needed
	assetByteOfs int64 // offset into chunk #0 where the asset data starts
	finalChunk   ChunkIdx
	track        *Spotify.Track
	trackFile    *Spotify.AudioFile
	downloader   *downloader
	cipher       cipher.Block
	decrypter    crypto.BlockDecrypter
	chunksMu     sync.Mutex
	chunks       map[ChunkIdx]*assetChunk
	//chunks        redblacktree.Tree

	residentChunkLimit int
	latestRead         ChunkIdx         // most recently requested chunk
	accessCount        int64            // incremented each time a different chunk is read
	fetchAhead         ChunkIdx         // number of chunks to fetch ahead of the current read position
	fetching           int32            // number of chunks currently being fetched
	onChunkComplete    chan *assetChunk // consumed by this assets runLoop to process completed chunks
	chunkChange        sync.Cond        // signaled when a chunk is available
	fatalErr           error
}

func newMediaAsset(dl *downloader, track *Spotify.Track) *mediaAsset {

	ma := &mediaAsset{
		downloader:      dl,
		track:           track,
		assetByteSz:     0,
		fetchAhead:      4,
		onChunkComplete: make(chan *assetChunk, 1),
		// chunks: redblacktree.Tree{
		// 	Comparator: func(A, B interface{}) int {
		// 		idx_a := A.(ChunkIdx)
		// 		idx_b := B.(ChunkIdx)
		// 		return int(idx_b - idx_a)
		// 	},
		// },

	}
	ma.chunkChange = sync.Cond{
		L: &ma.chunksMu,
	}

	// TEST ME
	ma.setResidentByteLimit(10 * 1024 * 1024)
	return ma
}

// needsFirstChunk returns if chunk 0 has been fetched.
// Once chunk 0 is fetched, this asset chunk profile is known.
func (a *mediaAsset) needsFirstChunk() bool {
	return a.assetByteSz == 0
}

func (a *mediaAsset) setResidentByteLimit(byteLimit int) {
	chunkLimit := 6 + byteLimit/kChunkByteSize
	a.residentChunkLimit = chunkLimit
	if a.chunks == nil {
		initialSz := min(chunkLimit, 70)
		a.chunks = make(map[ChunkIdx]*assetChunk, initialSz)
	}
}

func (ma *mediaAsset) Label() string {
	return ma.label
}

func (ma *mediaAsset) MediaType() string {
	return ma.mediaType
}

// pre: ma.chunksMu is locked
func (ma *mediaAsset) OnStart(ctx task.Context) error {
	ma.Context = ctx
	go ma.runLoop()
	return nil
}

// pre: ma.chunksMu is locked
func (ma *mediaAsset) getReadyChunk(idx ChunkIdx) *assetChunk {
	if chunk := ma.chunks[idx]; chunk != nil {
		if chunk.status == chunkReady {
			return chunk
		}
	}

	// val, exists := ma.chunks.Get(idx)
	// if exists {
	// 	chunk := val.(*assetChunk)
	// 	if chunk.status == chunkReady {
	// 		return chunk
	// 	}
	// }
	return nil
}

/*

// Steps rightward from an index until a hole is found (or the maxDelta is reached)
// pre: ma.chunksMu is locked
func (ma *mediaAsset) findNextChunkHole(startIdx, maxDelta ChunkIdx) ChunkIdx {
	holeIdx := ChunkIdx_Nil

	startNode := ma.chunks.GetNode(startIdx)
	if startNode == nil {
		holeIdx = startIdx
	} else {
		itr := ma.chunks.IteratorAt(startNode)
		for i := ChunkIdx(1); itr.Next() && i < maxDelta; i++ {
			idx := itr.Key().(ChunkIdx)
			idx_expected := startIdx + i
			if idx != idx_expected {
				holeIdx = idx_expected
				break
			}
		}
	}

	if holeIdx > ma.finalChunk {
		holeIdx = ChunkIdx_Nil
	}

	return holeIdx
}
*/

// func (ma *mediaAsset) checkErr(err error) {
// 	if err == nil {
// 		return
// 	}
// 	ma.throwErr(err)
// }

func (ma *mediaAsset) throwErr(err error) {
	if ma.fatalErr == nil {
		if err == nil {
			err = errors.New("unspecified fatal error")
		}
		ma.fatalErr = err
		ma.Context.Close()
	}
}

// Primary run loop for this media asset
// runs in its own goroutine and pushes events as they come
func (ma *mediaAsset) runLoop() {

	for running := true; running; {
		if ma.fatalErr != nil {
			return
		}

		select {
		case chunk := <-ma.onChunkComplete:
			if chunk.status == chunkReadyToDecrypt {
				ma.decrypter.DecryptSegment(chunk.Idx.StartByteOffset(), ma.cipher, chunk.Data, chunk.Data)
				chunk.status = chunkReady

				ma.chunksMu.Lock()
				ma.fetching -= 1
				if kDebug {
					ma.Context.Infof(2, "RECV chunk %d/%d/%d", chunk.Idx, ma.fetching, ma.finalChunk)
				}
				{
					// The size should only change when the first chunk arrives
					if sz := chunk.totalAssetSz; sz != ma.assetByteSz {
						ma.assetByteSz = sz
						ma.finalChunk = ChunkIdxAtOffset(sz)
					}
					ma.requestChunkIfNeeded(ma.latestRead)
				}
				ma.chunkChange.Broadcast()
				ma.chunksMu.Unlock()

			} else {
				panic("chunk failed")
			}

		case <-ma.Context.Closing():
			ma.throwErr(ma.Context.Err())
			running = false

			ma.chunksMu.Lock()
			ma.chunkChange.Broadcast()
			ma.chunksMu.Unlock()
		}
	}
}

// Pre: ma.chunksMu is locked
func (ma *mediaAsset) requestChunkIfNeeded(needIdx ChunkIdx) {

	fetchIdx := ChunkIdx_Nil

	// Find a chunk to fetch starting at the the chunk we need and step rightward
	{
		for ahead := ChunkIdx(0); ahead <= ma.fetchAhead; ahead++ {
			idx := needIdx + ahead
			if idx > ma.finalChunk {
				break
			}
			chunk := ma.chunks[idx]
			if chunk == nil {
				// Fetch only if not fetching ahead *and* within max concurrent fetches
				if ahead == 0 || ma.fetching < kMaxConcurrentFetches {
					fetchIdx = idx
				}
				break
			}
		}
	}

	if kDebug {
		ma.Infof(2, "requestChunkIfNeeded: needIdx: %d, fetchIdx: %d/%d/%d", needIdx, fetchIdx, ma.fetching, ma.finalChunk)
	}

	if fetchIdx >= 0 {
		chunk, err := ma.downloader.RequestChunk(fetchIdx, ma)
		if err != nil {
			ma.throwErr(err)
			return
		} else {
			ma.fetching += 1
			ma.chunks[fetchIdx] = chunk
		}
	}
	/*
		{
			L := readingAt
			R := readingAt + ma.fetchAhead
			if R > ma.finalChunk {
				R = ma.finalChunk
			}

			// find the next chunk we don't have withing the fetch ahead range
			for R - L > 0 {
				idx := (L + R) >> 1
				chunk := ma.chunks[idx]
				if chunk == nil {
					fetchIdx = idx
					R = idx
				} else {
					L = idx
				}
			}
			for ahead := ma.fetchAhead; ahead >= 0; ahead-- {
				idx := readingAt + ahead
				if idx > ma.finalChunk {
					break
				}
				chunk := ma.chunks[idx]
				if chunk == nil {
					fetchIdx = idx
				}
			}
		}

		// Update readyRange (bookkeeping)
		{
			idx := ma.readyRange.Start + ma.readyRange.Length
			for ; idx <= ma.finalChunk; idx++ {
				chunk := ma.getReadyChunk(idx)
				if chunk == nil {
					break
				}
				ma.readyRange.Length++  // readyRange bookkeeping
			}
		}


		// Reset ready range if the current read position is outside of it.
		// Include an additional rightmost element since
		if readingAt < ma.readyRange.Start || readingAt >= ma.readyRange.Start+ma.readyRange.Length {
			fetchIdx = readingAt
			ma.readyRange.Start = readingAt
			ma.readyRange.Length = 0
		}

		if ma.readyRange.Contains(readingAt) {
			readUpto := readingAt + ma.fetchAhead
			idx := ma.readyRange.Start + ma.readyRange.Length
			if idx <= ma.finalChunk && idx <= readUpto {
				chunk := ma.chunks[idx]
				if chunk == nil {
					fetchIdx = idx
				}
			}
		} else {
			fetchIdx = readingAt
			ma.readyRange.Start = readingAt
			ma.readyRange.Length = 0
		}

		// // Update readyRange (bookkeeping) and figure out what to fetch next
		// idx := ma.readyRange.Start + ma.readyRange.Length
		// for ; idx <= ma.finalChunk; idx++ {
		// 	chunk := ma.chunks[idx]
		// 	if chunk != nil {
		// 		if chunk.status == chunkReady {
		// 			ma.readyRange.Length++  // readyRange bookkeeping
		// 		}
		// 	} else {
		// 		if idx >= readingAt && idx <= readUpto {
		// 			fetchIdx = idx
		// 		}
		// 		break
		// 	}
		// }
	*/

}

// readChunk returns the chunk at the given index, blocking until it is available or a fatal error.
func (ma *mediaAsset) readChunk(idx ChunkIdx) (*assetChunk, error) {

	// Note: this is wired where chunk 0 is always assumed to exist (and is always the first accessed)
	if idx < 0 || idx > ma.finalChunk {
		return nil, io.EOF
	}

	ma.chunkChange.L.Lock()
	defer ma.chunkChange.L.Unlock()

	ma.latestRead = idx

	for ma.fatalErr == nil {

		// Call this before we exit to ensure we fetch ahead
		ma.requestChunkIfNeeded(idx)

		// Is the chunk ready? -- most of the time it will be since we read ahead
		chunk := ma.getReadyChunk(idx)
		if chunk != nil {
			ma.accessCount++
			chunk.accessStamp = ma.accessCount
			chunk.numReaders.Add(1)
			return chunk, nil
		}

		// Wait for signal; unlocks ma.chunkChange.L
		ma.chunkChange.Wait()
	}

	return nil, ma.fatalErr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (asset *mediaAsset) NewAssetReader() (arc.AssetReader, error) {
	reader := &assetReader{
		asset:   asset,
		readPos: 0,
	}
	return reader, nil
}

// func (ma *mediaAsset) headerOffset() int64 {
// 	// If the file format is an OGG, we skip the first kOggSkipBytes (167) bytes. We could implement despotify's
// 	// SpotifyOggHeader (https://sourceforge.net/p/despotify/code/HEAD/tree/java/trunk/src/main/java/se/despotify/client/player/SpotifyOggHeader.java)
// 	// to read Spotify's metadata (samples, length, gain, ...). For now, we simply skip the custom header to the actual
// 	// OGG/Vorbis data.
// 	return int64(ma.headerOfs)
// }

func (ci ChunkIdx) StartByteOffset() int64 {
	return kChunkByteSize * int64(ci)
}

func ChunkIdxAtOffset(byteOfs int64) ChunkIdx {
	return ChunkIdx(byteOfs >> 17) // int(math.Floor(float64(byteIndex) / float64(kChunkSize) / 4.0))
}

type assetReader struct {
	hotChunk atomic.Pointer[assetChunk]
	closed   atomic.Bool
	asset    *mediaAsset
	readPos  int64
}

// Read is an implementation of the io.Reader interface.
// This function will block until a non-zero amount of data is available (or io.EOF or a fatal error occurs).
func (r *assetReader) Read(buf []byte) (int, error) {
	if err := r.checkState(); err != nil {
		return 0, err
	}

	bytesRemain := len(buf)
	bytesRead := 0

	r.readPos = max(r.readPos, r.asset.assetByteOfs)

	if kDebug {
		pos := r.readPos - r.asset.assetByteOfs
		r.asset.Infof(2, "READ REQ %7d bytes, @%7d", bytesRemain, pos)
	}

	for bytesRemain > 0 {
		hotChunk, err := r.lockChunkAtOfs(r.readPos)
		if err != nil {
			return 0, err
		}

		relPos := int(r.readPos - hotChunk.Idx.StartByteOffset())
		runSz := len(hotChunk.Data) - relPos
		runSz = min(runSz, bytesRemain)
		if runSz > 0 {
			copy(buf[bytesRead:bytesRead+runSz], hotChunk.Data[relPos:relPos+runSz])
			r.readPos += int64(runSz)
			bytesRead += runSz
			bytesRemain -= runSz
		} else if runSz == 0 && hotChunk.Idx >= r.asset.finalChunk {
			if bytesRead == 0 {
				return 0, io.EOF
			}
			break
		} else {
			panic("bad runSz")
		}
	}

	if kDebug {
		pos := r.readPos - r.asset.assetByteOfs
		r.asset.Infof(2, " <--- COMPLETE %d bytes read, @%7d\n", bytesRead, pos)
	}

	return bytesRead, nil
}

func (r *assetReader) Seek(offset int64, whence int) (int64, error) {
	if err := r.checkState(); err != nil {
		return 0, err
	}

	switch whence {
	case io.SeekStart:
		r.readPos = offset + r.asset.assetByteOfs
	case io.SeekEnd:
		r.readPos = offset + r.asset.assetByteSz
	case io.SeekCurrent:
		r.readPos += offset
	}

	pos := r.readPos - r.asset.assetByteOfs
	if kDebug {
		r.asset.Context.Infof(2, "SEEK  %12d", pos)
	}
	return pos, nil
}

func (r *assetReader) checkState() error {
	if r.closed.Load() {
		return io.ErrClosedPipe
	}

	// We don't know anything until we have the first chunk
	if r.asset.needsFirstChunk() {
		_, err := r.lockChunkAtOfs(0)
		if err != nil {
			return err
		}
	}

	return nil
}

// gets and locks the returned the chunk containing the given starting data offset, blocking until it is available (or fatal error).
func (r *assetReader) lockChunkAtOfs(rawOfs int64) (*assetChunk, error) {
	idx := ChunkIdxAtOffset(rawOfs)

	// If we already have the chunk, no need to get it from the asset.
	{
		hotChunk := r.hotChunk.Load()
		if hotChunk != nil && hotChunk.Idx == idx {
			return hotChunk, nil
		}
	}

	chunk, err := r.asset.readChunk(idx)
	if err == nil {
		err = r.lockChunk(chunk)
	}
	return chunk, err
}

func (r *assetReader) lockChunk(chunk *assetChunk) error {
	var err error
	
	// If this reader was closed, release the chunk we want to lock
	if chunk != nil && r.closed.Load() {
		chunk.numReaders.Add(-1)
		chunk = nil
		err = io.ErrClosedPipe
	}
	
	// Release the prev chunk we locked if applicable.
	prev := r.hotChunk.Swap(chunk)
	if prev != nil {
		prev.numReaders.Add(-1)
	}
	return err
}

func (r *assetReader) Close() error {
	if r.closed.CompareAndSwap(false, true) {
		r.lockChunk(nil)
		if kDebug {
			r.asset.Infof(2, "CLOSED")
		}
	}
	return nil
}
