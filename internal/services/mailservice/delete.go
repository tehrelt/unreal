package mailservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *Service) Delete(ctx context.Context, mailbox string, num uint32) error {
	fn := "mailservice.Delete"
	log := s.l.With(sl.Method(fn), slog.String("mailbox", mailbox), slog.Int("num", int(num)))

	if err := s.m.Do(ctx, func(ctx context.Context) (err error) {
		if err := s.r.Delete(ctx, mailbox, num); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Error("cannot delete message")
		return fmt.Errorf("%s: %v", fn, err)
	}

	return nil
}
