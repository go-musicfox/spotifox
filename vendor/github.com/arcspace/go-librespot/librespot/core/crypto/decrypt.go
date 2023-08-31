package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"math/big"
)

var AUDIO_AESIV = []byte{0x72, 0xe0, 0x67, 0xfb, 0xdd, 0xcb, 0xcf, 0x77, 0xeb, 0xe8, 0xbc, 0x64, 0x3f, 0x63, 0x0d, 0x93}

type BlockDecrypter struct {
	ivDiff big.Int
	ivInt  big.Int
}

func CreateCipher(key []byte) cipher.Block {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	return block
}

func (afd *BlockDecrypter) DecryptSegment(byteOfs int64, block cipher.Block, ciphertext []byte, plaintext []byte) {
	length := len(ciphertext)
	// plaintext := bufferPool.Get().([]byte) // make([]byte, length)

	// The actual IV is the base IV + index*0x100, where index is the chunk index sized 1024 words (so each 4096 bytes
	// block has its own IV). As we are retrieving 32768 words (131072 bytes) to speed up network operations, we need
	// to process the data by 4096 bytes blocks to decrypt with the correct key.

	// We pre-calculate the base IV for the first chunk we are processing, then just proceed to add 0x100 at
	// every iteration.
	afd.ivInt.SetBytes(AUDIO_AESIV)
	afd.ivDiff.SetInt64((byteOfs >> 12) << 8)
	afd.ivInt.Add(&afd.ivInt, &afd.ivDiff)

	afd.ivDiff.SetInt64(int64(0x100))

	for i := 0; i < length; i += 4096 {
		i_end := i + 4096
		if i_end > length {
			i_end = length
		}
		stream := cipher.NewCTR(block, afd.ivInt.Bytes())
		stream.XORKeyStream(plaintext[i:i_end], ciphertext[i:i_end])
		afd.ivInt.Add(&afd.ivInt, &afd.ivDiff)
	}
}
