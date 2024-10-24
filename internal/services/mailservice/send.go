package mailservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *Service) Send(ctx context.Context, req *dto.SendMessageDto) error {

	fn := "mailservice.Send"
	log := s.l.With(sl.Method(fn))

	if err := s.m.Do(ctx, func(ctx context.Context) error {

		in := &models.SendMessage{
			From:        req.From,
			To:          req.To,
			Body:        req.Body,
			Subject:     req.Subject,
			Attachments: req.Attachments,
		}

		u, err := s.userProvider.Find(ctx, in.From.Email)
		if err != nil {
			if !errors.Is(err, storage.ErrUserNotFound) {
				log.Warn("failed to find user", sl.Err(err))
				return fmt.Errorf("%s: %w", fn, err)
			}
		}
		in.From.SetName(u.Name)

		done := make(chan error)
		go func(ctx context.Context) {
			in.To = lo.Map(in.To, func(r *dto.MailRecord, _ int) *dto.MailRecord {
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
			id := uuid.NewString()

			in.Body, err = s.encrypt(ctx, id, req.Body)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			in.EncryptKey = &id
		}

		msg, err := s.sender.Send(ctx, in)
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
