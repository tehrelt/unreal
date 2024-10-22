package mailservice

import (
	"context"
	"fmt"
	"io"
)

func (s *Service) Attachment(ctx context.Context, mailbox string, mailnum uint32, target string) (r io.Reader, ct string, err error) {
	fn := "mailservice.GetAttachment"

	if err := s.m.Do(ctx, func(ctx context.Context) error {
		r, ct, err = s.r.Attachment(ctx, mailbox, mailnum, target)
		return err
	}); err != nil {
		return nil, "", fmt.Errorf("%s: %w", fn, err)
	}

	return
}
