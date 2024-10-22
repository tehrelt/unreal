package mailservice

import (
	"context"
	"fmt"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *Service) Send(ctx context.Context, req *dto.SendMessageDto) error {

	fn := "mailservice.Send"
	log := s.l.With(sl.Method(fn))

	if err := s.m.Do(ctx, func(ctx context.Context) error {

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
