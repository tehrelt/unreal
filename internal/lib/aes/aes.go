package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
)

type AesEncryptor struct {
	secretkey string
}

func NewAesEncryptor(secretkey string) *AesEncryptor {
	return &AesEncryptor{secretkey: secretkey}
}

func (e *AesEncryptor) Encrypt(plaintext string) (string, error) {

	slog.Debug("encrypting", slog.String("secretkey", e.secretkey))
	key := []byte(e.secretkey)
	text := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %v", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (e *AesEncryptor) Decrypt(toDecrypt string) (string, error) {

	key := []byte(e.secretkey)
	ciphertext, err := base64.URLEncoding.DecodeString(toDecrypt)
	if err != nil {
		return "", fmt.Errorf("base64.URLEncoding.DecodeString: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %v", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
