package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"math"
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

func (c *Cipher) Encrypt(in []byte) ([]byte, error) {
	in = pad(in, c.blockSize)
	l := len(in)
	out := make([]byte, 0, l)

	iters := int(math.Ceil(float64(l) / float64(c.blockSize)))
	for i := 0; i < iters; i++ {
		start := i * c.blockSize
		end := (i + 1) * c.blockSize
		if end > l {
			end = l
		}

		enc, err := c.encrypt(in[start:end])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", "Encrypt", err)
		}

		out = append(out, enc...)
	}

	return out, nil
}

func (c *Cipher) Decrypt(in []byte) ([]byte, error) {
	l := len(in)
	out := make([]byte, 0, l)

	iters := l / c.blockSize
	if iters%2 != 0 {
		return nil, fmt.Errorf("malformed input")
	}

	for i := 0; i < iters; i += 2 {
		start := i * c.blockSize
		end := (i + 1) * c.blockSize
		iv := in[start:end]

		start = end
		end = (i + 2) * c.blockSize
		data := in[start:end]

		dec := c.decrypt(data, iv)
		out = append(out, dec...)
	}

	out = unpad(out)
	return out, nil
}

func (c *Cipher) encrypt(input []byte) ([]byte, error) {
	output := make([]byte, len(input)+c.blockSize)

	iv := output[:c.blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	encrypter := cipher.NewCBCEncrypter(c.block, iv)
	encrypter.CryptBlocks(output[c.blockSize:], input)

	return output, nil
}

func (c *Cipher) decrypt(input, iv []byte) []byte {
	block := cipher.NewCBCDecrypter(c.block, iv)
	output := make([]byte, len(input))
	block.CryptBlocks(output, input)
	return output
}

func pad(in []byte, blockSize int) []byte {
	padding := blockSize - len(in)%blockSize
	if padding == blockSize {
		padding = 0
	}
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(in, padText...)
}

func unpad(in []byte) []byte {
	length := len(in)
	if length == 0 {
		return in
	}

	unpadding := int(in[length-1])
	if unpadding < 16 {
		return in[:length-unpadding]
	}

	return in
}
