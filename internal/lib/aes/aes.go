package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type AesEncryptor struct {
	secretkey []byte
}

func NewAesEncryptor(secretkey []byte) *AesEncryptor {
	return &AesEncryptor{secretkey: secretkey}
}

func (e *AesEncryptor) Encrypt(plaintext []byte) ([]byte, error) {

	fn := "aes.Encrypt"
	// log := slog.With(sl.Method(fn))

	key := e.secretkey
	text := plaintext

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return ciphertext, nil
}

func (e *AesEncryptor) Decrypt(toDecrypt []byte) ([]byte, error) {

	fn := "aes.Decrypt"

	key := []byte(e.secretkey)
	ciphertext := toDecrypt

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", fn, err)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
