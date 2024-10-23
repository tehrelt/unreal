package aes

import "encoding/base64"

type StringCipher struct {
	e *AesEncryptor
}

func NewStringCipher(secretkey []byte) *StringCipher {
	return &StringCipher{
		e: NewAesEncryptor(secretkey),
	}
}

func (c *StringCipher) Encrypt(data string) (string, error) {
	b := []byte(data)

	s, err := c.e.Encrypt(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(s), nil
}

func (c *StringCipher) Decrypt(data string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	s, err := c.e.Decrypt(b)
	if err != nil {
		return "", err
	}

	return string(s), nil
}
