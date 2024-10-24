package mailservice

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/tehrelt/unreal/internal/lib/aes"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *Service) encrypt(ctx context.Context, id string, body io.Reader) (io.Reader, error) {

	fn := "mailservice.encrypt"
	log := s.l.With(sl.Method(fn))

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Debug("failed to generate random key", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	encryptor, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	hashsum := sha1.Sum(data)
	encdata, err := encryptor.Encrypt(data)
	if err != nil {
		log.Debug("failed to encrypt message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	log.Debug("encrypted message", slog.Any("original", data), slog.Any("data", encdata))

	enckey, err := s.keyCipher.Encrypt(bytes.NewReader(key))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	ekey, err := io.ReadAll(enckey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	in := &models.VaultRecord{
		Id:      id,
		Key:     base64.StdEncoding.EncodeToString(ekey),
		Hashsum: base64.StdEncoding.EncodeToString(hashsum[:]),
	}
	if err := s.vault.Insert(ctx, in); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return io.NopCloser(bytes.NewReader(encdata)), nil
}

func (s *Service) decrypt(ctx context.Context, id string, body io.Reader) (io.Reader, error) {
	fn := "mailservice.decrypt"
	log := s.l.With(sl.Method(fn))

	rec, err := s.vault.Find(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	key, err := base64.StdEncoding.DecodeString(rec.Key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	origkey, err := s.keyCipher.Decrypt(bytes.NewReader(key))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	key, err = io.ReadAll(origkey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	decryptor, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	decdata, err := decryptor.Decrypt(data)
	if err != nil {
		log.Debug("failed to decrypt message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	log.Debug("decrypted message", slog.String("data", string(decdata)))

	actualsum := sha1.Sum(decdata)
	expectedsum, err := base64.StdEncoding.DecodeString(rec.Hashsum)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if !bytes.Equal(actualsum[:], expectedsum[:]) {
		return nil, fmt.Errorf("%s: %w", fn, errors.New("hashsum does not match"))
	}
	log.Info("hashsum matches")

	return io.NopCloser(bytes.NewReader(decdata)), nil
}
