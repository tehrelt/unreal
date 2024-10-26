package mailservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/storage"
	"github.com/tehrelt/unreal/internal/storage/models"
)

func (s *Service) Send(ctx context.Context, req *dto.SendMessageDto) error {

	fn := "mailservice.Send"
	log := s.l.With(sl.Method(fn))

	if err := s.m.Do(ctx, func(ctx context.Context) error {

		in := &models.SendMessage{
			From:        req.From,
			To:          req.To,
			Subject:     req.Subject,
			Body:        req.Body,
			Attachments: req.Attachments,
		}

		u, err := s.userProvider.Find(ctx, in.From.Email)
		if err != nil {
			if !errors.Is(err, storage.ErrUserNotFound) {
				log.Warn("failed to find user", sl.Err(err))
				return fmt.Errorf("%s: %w", fn, err)
			}
		}
		in.From.SetName(u.Name)

		done := make(chan error)
		go func(ctx context.Context) {
			in.To = lo.Map(in.To, func(r *dto.MailRecord, _ int) *dto.MailRecord {
				if err != nil {
					return nil
				}

				u, err := s.userProvider.Find(ctx, r.Email)
				if err != nil {
					if !errors.Is(err, storage.ErrUserNotFound) {
						log.Warn("failed to find user", sl.Err(err))
						done <- err
						return nil
					}
				}

				if u != nil {
					r.SetName(u.Name)
				}

				return r
			})

			done <- nil
		}(ctx)

		err = <-done
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		if req.DoEncryiption {
			id := uuid.NewString()

			buf := new(bytes.Buffer)
			enc, err := s.encrypt(ctx, id, req.Body)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			fwr := multipart.NewWriter(buf)

			fpt, err := fwr.CreateFormFile("file", ".unreal")
			if err != nil {
				log.Error("cannot create form file", sl.Err(err))
				return fmt.Errorf("%s: %w", fn, err)
			}

			if _, err := io.Copy(fpt, enc); err != nil {
				log.Error("cannot copy encoded data to form part", sl.Err(err))
				return fmt.Errorf("%s: %w", fn, err)
			}
			fwr.Close()

			bufreader := bytes.NewReader(buf.Bytes())
			frd := multipart.NewReader(bufreader, fwr.Boundary())

			frm, err := frd.ReadForm(1 << 20)
			if err != nil {
				log.Error("cannot read form", sl.Err(err))
				return fmt.Errorf("%s: %w", fn, err)
			}
			in.Attachments = append(in.Attachments, frm.File["file"][0])

			in.Body = new(bytes.Buffer)

			in.EncryptKey = &id
		}

		msg, err := s.sender.Send(ctx, in)
		if err != nil {
			log.Error("cannot send message", sl.Err(err))
			return fmt.Errorf("%s: %w", fn, err)
		}

		if err := s.r.SaveMessageToFolderByAttribute(ctx, "\\Sent", msg); err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		return nil
	}); err != nil {
		log.Error("cannot send message", sl.Err(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
