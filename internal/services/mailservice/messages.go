package mailservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services"
	"github.com/tehrelt/unreal/internal/storage"
)

const (
	defaultLimit = 50
)

func (s *Service) Messages(ctx context.Context, in *dto.FetchMessagesDto) (*dto.FetchedMessagesDto, error) {
	fn := "mailservice.Messages"
	log := slog.With(sl.Method(fn))

	var out *dto.FetchedMessagesDto

	if err := s.m.Do(ctx, func(ctx context.Context) error {
		var err error

		log.Debug(
			"fetching messages",
			slog.Int("page", in.Page),
			slog.Int("limit", in.Limit),
			slog.String("mailbox", in.Mailbox),
		)
		out, err = s.r.Messages(ctx, in)
		if err != nil {
			return err
		}

		for i, m := range out.Messages {
			addr := m.From.Address
			u, err := s.userProvider.Find(ctx, addr)
			if err != nil {
				if !errors.Is(err, storage.ErrUserNotFound) {
					return fmt.Errorf("%s: %w", fn, err)
				}
			}

			if u != nil {
				log.Debug("found user", slog.Any("user", u))
				if u.ProfilePicture != nil {
					link := services.GetPictureLink(s.cfg.Host(), *u.ProfilePicture)
					log.Debug("profile picture link", slog.String("link", link))
					out.Messages[i].From.Picture = link
				}
			}

		}

		return nil
	}); err != nil {
		log.Error("failed to fetch messages", sl.Err(err))
		return nil, err
	}

	return out, nil
}
