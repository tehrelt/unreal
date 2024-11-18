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

type encrypted struct {
	data []byte
	r    io.Reader
	l    int64
	k    string
	sum  string
}
type decrypted struct {
	r io.Reader
	l int64
}

func (s *Service) encryptBody(ctx context.Context, mid string, body io.Reader) (*encrypted, error) {

	fn := "mailservice.encryptBody"

	enc, err := s.encrypt(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	in := &models.VaultRecord{
		Id:      mid,
		Key:     enc.k,
		Hashsum: enc.sum,
	}
	if err := s.vault.Insert(ctx, in); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return enc, nil
}

func (s *Service) encrypt(body io.Reader) (*encrypted, error) {

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

	out := &encrypted{
		data: encdata,
		r:    io.NopCloser(bytes.NewReader(encdata)),
		l:    int64(len(encdata)),
		k:    base64.StdEncoding.EncodeToString(ekey),
		sum:  base64.StdEncoding.EncodeToString(hashsum[:]),
	}

	return out, nil
}

func (s *Service) decryptBody(ctx context.Context, id string, body io.Reader) (*decrypted, error) {
	fn := "mailservice.decryptBody"

	rec, err := s.vault.Find(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	key, err := base64.StdEncoding.DecodeString(rec.Key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	hashsum, err := base64.StdEncoding.DecodeString(rec.Hashsum)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return s.decrypt(key, hashsum, body)
}

func (s *Service) decryptFile(ctx context.Context, id string, body io.Reader) (*decrypted, error) {
	fn := "mailservice.decryptFile"

	rec, err := s.vault.FileById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	key, err := base64.StdEncoding.DecodeString(rec.Key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	hashsum, err := base64.StdEncoding.DecodeString(rec.Hashsum)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return s.decrypt(key, hashsum, body)
}

func (s *Service) decrypt(key, hashsum []byte, body io.Reader) (*decrypted, error) {
	fn := "mailservice.decrypt"
	log := s.l.With(sl.Method(fn))

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
	ciphertext, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	plaintext, err := decryptor.Decrypt(ciphertext)
	if err != nil {
		log.Debug("failed to decrypt message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	log.Debug("decrypted message", slog.Any("ciphertext", ciphertext), slog.Any("plaintext", plaintext))

	actualsum := sha1.Sum(plaintext)
	if !bytes.Equal(actualsum[:], hashsum) {
		log.Error("hashsum does not match", slog.Any("actual", actualsum[:]), slog.Any("expected", hashsum))
		return nil, fmt.Errorf("%s: %w", fn, errors.New("hashsum does not match"))
	}
	log.Info("hashsum matches")

	out := &decrypted{
		r: io.NopCloser(bytes.NewReader(plaintext)),
		l: int64(len(plaintext)),
	}

	return out, nil
}
