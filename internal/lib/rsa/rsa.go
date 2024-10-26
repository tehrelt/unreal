package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"io"
)

type Cipher struct {
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

func New(priv *rsa.PrivateKey, pub *rsa.PublicKey) *Cipher {
	return &Cipher{
		privKey: priv,
		pubKey:  pub,
	}
}

func (c *Cipher) Encrypt(in io.Reader) (io.Reader, error) {
	if c.pubKey == nil {
		return nil, errors.New("public key is required for encryption")
	}

	data, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	ciphertext, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, c.pubKey, data, nil)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(ciphertext)), nil
}

func (c *Cipher) Decrypt(in io.Reader) (io.Reader, error) {
	if c.privKey == nil {
		return nil, errors.New("private key is required for decryption")
	}

	ciphertext, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	plaintext, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, c.privKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(plaintext)), nil
}
