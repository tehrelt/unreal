package mailservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
			messageId := uuid.NewString()

			args := &models.AppendFilesArgs{
				VaultFileBase: models.VaultFileBase{
					MessageId: messageId,
				},
				Files: make([]models.VaultFileMeta, 0, len(in.Attachments)),
			}

			for i, f := range in.Attachments {
				log.Debug("encrypting attachment", slog.String("filename", f.Filename))
				body, err := f.Open()
				if err != nil {
					return fmt.Errorf("%s: %w", fn, err)
				}
				defer body.Close()

				enc, err := s.encrypt(body)
				if err != nil {
					return fmt.Errorf("%s: %w", fn, err)
				}

				fid := uuid.NewString()
				args.Files = append(args.Files, models.VaultFileMeta{
					FileId:      fid,
					Filename:    f.Filename,
					ContentType: f.Header.Get("Content-Type"),
					Key:         enc.k,
					Hashsum:     enc.sum,
				})

				f, err := createMultipartHeader(fid, enc.r, enc.l)
				if err != nil {
					return fmt.Errorf("%s: %w", fn, err)
				}

				in.Attachments[i] = f
			}

			enc, err := s.encryptBody(ctx, messageId, req.Body)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}
			if len(in.Attachments) != 0 {
				if err := s.vault.AppendFiles(ctx, args); err != nil {
					return fmt.Errorf("%s: %w", fn, err)
				}
			}
			body, err := createMultipartHeader(".unreal", enc.r, enc.l)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}
			in.Attachments = append(in.Attachments, body)

			in.Body = new(bytes.Buffer)
			in.EncryptKey = messageId
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

func createMultipartHeader(name string, r io.Reader, size int64) (*multipart.FileHeader, error) {

	fn := "createMultipartHeader"
	log := slog.With(sl.Method(fn))

	buf := new(bytes.Buffer)
	fwr := multipart.NewWriter(buf)

	fpt, err := fwr.CreateFormFile("file", name)
	if err != nil {
		log.Error("cannot create form file", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	if _, err := io.Copy(fpt, r); err != nil {
		log.Error("cannot copy encoded data to form part", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	fwr.Close()

	bufreader := bytes.NewReader(buf.Bytes())
	frd := multipart.NewReader(bufreader, fwr.Boundary())

	frm, err := frd.ReadForm(size)
	if err != nil {
		log.Error("cannot read form", sl.Err(err))
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return frm.File["file"][0], nil
}
