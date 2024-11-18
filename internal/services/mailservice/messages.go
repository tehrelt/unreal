package mailservice

import (
	"context"
	"log/slog"
	"sync"

	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
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

		wg := &sync.WaitGroup{}
		wg.Add(len(out.Messages))
		for i, m := range out.Messages {
			go func(i int) {
				defer wg.Done()

				out.Messages[i].From, err = s.fillAddressInfo(ctx, m.From)
				if err != nil {
					log.Error("cannot fetch picture of sender", sl.Err(err))
				}
			}(i)
		}

		wg.Wait()

		return nil
	}); err != nil {
		log.Error("failed to fetch messages", sl.Err(err))
		return nil, err
	}

	return out, nil
}
