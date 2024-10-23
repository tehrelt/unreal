package mailservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
)

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

				u, nerr := s.userProvider.Find(ctx, r.Email)
				if nerr != nil {
					if !errors.Is(err, storage.ErrUserNotFound) {
						log.Warn("failed to find user", sl.Err(err))
						done <- err
						return nil
					}
				}

				r.SetName(u.Name)

				return r
			})

			done <- nil
		}(ctx)

		err = <-done
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
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
