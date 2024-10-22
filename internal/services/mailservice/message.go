package mailservice

import (
	"context"
	"log/slog"

	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *MailService) Message(ctx context.Context, mailbox string, num uint32) (out *entity.MessageWithBody, err error) {
	fn := "mailservice.Message"
	log := s.l.With(sl.Method(fn), slog.String("mailbox", mailbox), slog.Int("num", int(num)))

	if err := s.m.Do(ctx, func(ctx context.Context) (err error) {
		out, err = s.r.Message(ctx, mailbox, num)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Error("cannot fetch message")
		return nil, err
	}

	return out, nil
}
