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

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/aes"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *Service) encrypt(ctx context.Context, body io.Reader) (io.Reader, error) {

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

	encryptor := aes.NewAesEncryptor(key)
	hashsum := sha1.Sum(data)
	encdata, err := encryptor.Encrypt(data)
	if err != nil {
		log.Debug("failed to encrypt message", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	enckey, err := s.keyCipher.Encrypt(bytes.NewReader(key))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	ekey, err := io.ReadAll(enckey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	in := &models.VaultRecord{
		Id:      uuid.NewString(),
		Key:     base64.StdEncoding.EncodeToString(ekey),
		Hashsum: base64.StdEncoding.EncodeToString(hashsum[:]),
	}
	if err := s.vault.Insert(ctx, in); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return io.NopCloser(bytes.NewReader(encdata)), nil
}

func (s *Service) Send(ctx context.Context, req *dto.SendMessageDto) error {

	fn := "mailservice.Send"
	log := s.l.With(sl.Method(fn))

	if err := s.m.Do(ctx, func(ctx context.Context) error {

		u, err := s.userProvider.Find(ctx, req.From.Email)
		if err != nil {
			if !errors.Is(err, storage.ErrUserNotFound) {
				log.Warn("failed to find user", sl.Err(err))
				return fmt.Errorf("%s: %w", fn, err)
			}
		}
		req.From.SetName(u.Name)

		done := make(chan error)
		go func(ctx context.Context) {
			req.To = lo.Map(req.To, func(r *dto.MailRecord, _ int) *dto.MailRecord {
				if err != nil {
					return nil
				}

				u, err := s.userProvider.Find(ctx, r.Email)
				if err != nil {
					if !errors.Is(err, storage.ErrUserNotFound) {
						log.Warn("failed to find user", sl.Err(err))
						done <- err
						return nil
					}
				}

				if u != nil {
					r.SetName(u.Name)
				}

				return r
			})

			done <- nil
		}(ctx)

		err = <-done
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		if req.DoEncryiption {
			req.Body, err = s.encrypt(ctx, req.Body)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}
		}

		msg, err := s.sender.Send(ctx, req)
		if err != nil {
			log.Error("cannot send message", sl.Err(err))
			return fmt.Errorf("%s: %w", fn, err)
		}

		if err := s.r.SaveMessageToFolderByAttribute(ctx, "\\Sent", msg); err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
