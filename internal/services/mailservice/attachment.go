package mailservice

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"

	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func (s *Service) Attachment(ctx context.Context, mailbox string, mailnum uint32, target string) (r io.Reader, ct string, err error) {
	fn := "mailservice.Attachment"
	log := s.l.With(sl.Method(fn))

	if err := s.m.Do(ctx, func(ctx context.Context) error {

		vaultId, err := s.r.IsMessageEncrypted(ctx, mailbox, mailnum)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		log.Debug("message is encrypted?", slog.String("vaultId", vaultId))
		if vaultId != "" {
			rec, err := s.vault.File(ctx, vaultId, target)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}
			log.Debug("file found in vault", slog.String("fileId", rec.FileId), slog.String("filename", rec.Filename))

			attachment, err := s.r.Attachment(ctx, mailbox, mailnum, rec.FileId)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			key, err := base64.StdEncoding.DecodeString(rec.Key)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			sum, err := base64.StdEncoding.DecodeString(rec.Hashsum)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			dec, err := s.decrypt(key, sum, attachment.R)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			r = dec.r
			ct = rec.ContentType
		} else {
			attachment, err := s.r.Attachment(ctx, mailbox, mailnum, target)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			r = attachment.R
			ct = attachment.ContentType
		}

		return nil
	}); err != nil {
		return nil, "", fmt.Errorf("%s: %w", fn, err)
	}

	return
}
