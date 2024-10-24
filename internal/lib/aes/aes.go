package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type Cipher struct {
	key       []byte
	block     cipher.Block
	blockSize int
}

func NewCipher(key []byte) (e *Cipher, err error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("key must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &Cipher{
		key:       key,
		block:     block,
		blockSize: aes.BlockSize,
	}, nil
}

func (c *Cipher) Encrypt(input []byte) ([]byte, error) {
	nonce := make([]byte, c.blockSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(c.block, nonce)
	out := &cipher.StreamReader{S: stream, R: bytes.NewReader(input)}

	output := make([]byte, c.blockSize)
	copy(output, nonce)

	r := io.MultiReader(bytes.NewReader(output), out)
	return io.ReadAll(r)
}

func (c *Cipher) Decrypt(input []byte) ([]byte, error) {
	nonce := input[:c.blockSize]

	stream := cipher.NewCFBDecrypter(c.block, nonce)
	out := &cipher.StreamReader{S: stream, R: bytes.NewReader(input[c.blockSize:])}
	return io.ReadAll(out)
}
