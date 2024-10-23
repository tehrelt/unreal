package mailservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services"
	"github.com/tehrelt/unreal/internal/storage"
)

func (s *Service) getProfilePicture(ctx context.Context, r entity.AddressInfo) (out entity.AddressInfo, err error) {

	fn := "mailservice.getUser"

	u, err := s.userProvider.Find(ctx, r.Address)
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			return out, fmt.Errorf("%s: %w", fn, err)
		}
	}

	if u != nil {
		if u.ProfilePicture != nil {
			r.Picture = services.GetPictureLink(s.cfg.Host(), *u.ProfilePicture)
		}
	}

	return r, nil
}

func (s *Service) Message(ctx context.Context, mailbox string, num uint32) (out *entity.MessageWithBody, err error) {
	fn := "mailservice.Message"
	log := s.l.With(sl.Method(fn), slog.String("mailbox", mailbox), slog.Int("num", int(num)))

	if err := s.m.Do(ctx, func(ctx context.Context) (err error) {
		out, err = s.r.Message(ctx, mailbox, num)
		if err != nil {
			return err
		}

		out.From, err = s.getProfilePicture(ctx, out.From)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		for i := range out.To {
			out.To[i], err = s.getProfilePicture(ctx, out.To[i])
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}
		}

		return nil
	}); err != nil {
		log.Error("cannot fetch message")
		return nil, err
	}

	return out, nil
}
