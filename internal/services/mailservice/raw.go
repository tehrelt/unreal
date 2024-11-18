package mailservice

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *Service) Raw(ctx context.Context, mailbox string, num uint32) (r io.Reader, err error) {
	fn := "mailservice.Raw"
	log := s.l.With(sl.Method(fn), slog.String("mailbox", mailbox), slog.Int("num", int(num)))

	if err := s.m.Do(ctx, func(ctx context.Context) error {

		r, err = s.r.Raw(ctx, mailbox, num)
		if err != nil {
			log.Error("cannot fetch body")
			return fmt.Errorf("%s: %w", fn, err)
		}

		return nil
	}); err != nil {
		log.Error("cannot fetch raw message")
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return r, nil
}
