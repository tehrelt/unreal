package mailservice

import (
	"context"
	"log/slog"

	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *MailService) Mailboxes(ctx context.Context) ([]*entity.Mailbox, error) {

	fn := "mailservice.Mailboxes"
	log := slog.With(sl.Method(fn))

	var mailboxes []*entity.Mailbox

	if err := s.m.Do(ctx, func(ctx context.Context) error {
		var err error

		log.Debug("list mailboxes")
		mailboxes, err = s.r.Mailboxes(ctx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return mailboxes, nil
}
