package aes_test

import (
	"crypto/rand"
	"crypto/sha1"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tehrelt/unreal/internal/lib/aes"
)

func TestAES(t *testing.T) {
	t.Log("aes_test")

	cases := []struct {
		name   string
		phrase string
	}{
		{"lower then 16 bytes", "hello world"},
		{"bigger then 16 bytes", "hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world"},
		{"empty", ""},
		{"multiline", `
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
WARNING: Panel defaultSize prop recommended to avoid layout shift after server rendering
		`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key := make([]byte, 32)
			_, err := io.ReadFull(rand.Reader, key)
			require.NoError(t, err)

			cipher, err := aes.NewCipher(key)
			require.NoError(t, err)

			plaintext := []byte(c.phrase)
			ciphertext, err := cipher.Encrypt(plaintext)
			require.NoError(t, err)

			original, err := cipher.Decrypt(ciphertext)
			require.NoError(t, err)

			require.Equal(t, plaintext, original)

			expectedsum := sha1.Sum([]byte(c.phrase))
			actualsum := sha1.Sum(original)
			require.Equal(t, expectedsum, actualsum)
		})
	}
}
