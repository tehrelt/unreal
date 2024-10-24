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
	input = pad(input, c.blockSize)
	output := make([]byte, len(input)+c.blockSize)

	iv := output[:c.blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	encrypter := cipher.NewCBCEncrypter(c.block, iv)
	encrypter.CryptBlocks(output[c.blockSize:], input)

	return output, nil
}

func (c *Cipher) Decrypt(input []byte) ([]byte, error) {
	iv := input[:c.blockSize]
	block := cipher.NewCBCDecrypter(c.block, iv)
	input = input[c.blockSize:]
	block.CryptBlocks(input, input)
	output := unpad(input)

	return output, nil
}

func pad(in []byte, blockSize int) []byte {
	padding := blockSize - len(in)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(in, padText...)
}

func unpad(in []byte) []byte {
	length := len(in)
	unpadding := int(in[length-1])
	return in[:length-unpadding]
}
