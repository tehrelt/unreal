package mailservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/entity"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
	"github.com/tehrelt/unreal/internal/services"
	"github.com/tehrelt/unreal/internal/storage"
)

func domain(addr string) string {
	return strings.Split(addr, "@")[1]
}

func (s *Service) fillAddressInfo(ctx context.Context, r entity.AddressInfo) (out entity.AddressInfo, err error) {

	fn := "mailservice.getUser"
	domain := domain(r.Address)

	pic, err := s.hostProvider.Find(ctx, domain)
	if err != nil {
		if !errors.Is(err, storage.ErrHostNotFound) {
			return out, fmt.Errorf("%s: %w", fn, err)
		}
	}

	if pic != "" {
		r.Picture = services.GetPictureLink(pic)
	}

	u, err := s.userProvider.Find(ctx, r.Address)
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			return out, fmt.Errorf("%s: %w", fn, err)
		}
	}

	if u != nil {
		if u.Name != nil {
			r.Name = *u.Name
		}

		if u.ProfilePicture != nil {
			r.Picture = services.GetPictureLink(*u.ProfilePicture)
		}
	}

	return r, nil
}

func (s *Service) Message(ctx context.Context, mailbox string, num uint32) (*entity.MessageWithBody, error) {
	fn := "mailservice.Message"
	log := s.l.With(sl.Method(fn), slog.String("mailbox", mailbox), slog.Int("num", int(num)))

	var out *entity.MessageWithBody

	if err := s.m.Do(ctx, func(ctx context.Context) (err error) {
		msg, err := s.r.Message(ctx, mailbox, num)
		if err != nil {
			return err
		}

		if msg.VaultId != "" {

			encfile, found := lo.Find(msg.Attachments, func(a entity.Attachment) bool {
				return a.Filename == ".unreal"
			})

			if found {
				attach, err := s.r.Attachment(ctx, mailbox, num, encfile.Filename)
				if err != nil {
					log.Error("failed to fetch attachment with encoded html")
					return fmt.Errorf("%s: %w", fn, err)
				}

				msg.Body = attach.R

				msg.Attachments = lo.Filter(msg.Attachments, func(a entity.Attachment, _ int) bool {
					return a.Filename != ".unreal"
				})
			}

			body, err := io.ReadAll(msg.Body)
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			if msg.Sign != "" {
				signature, err := base64.StdEncoding.DecodeString(msg.Sign)
				if err != nil {
					log.Error("failed to decode signature")
					return fmt.Errorf("%s: %w", fn, err)
				}

				if err := s.signer.Verify(body, signature); err != nil {
					log.Error("failed to verify signature")
					return fmt.Errorf("%s: %w", fn, err)
				}

				log.Info("successfully verified signature")
			} else {
				log.Warn("no signature on ciphered message")
			}

			dec, err := s.decryptBody(ctx, msg.VaultId, bytes.NewBuffer(body))
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			msg.Body = dec.r

			for i, f := range msg.Attachments {
				rec, err := s.vault.FileById(ctx, f.Filename)
				if err != nil {
					return fmt.Errorf("%s: %w", fn, err)
				}

				msg.Attachments[i] = entity.Attachment{
					ContentId:   rec.Filename,
					Filename:    rec.Filename,
					ContentType: rec.ContentType,
				}
			}
		}

		msg.From, err = s.fillAddressInfo(ctx, msg.From)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		for i := range msg.To {
			msg.To[i], err = s.fillAddressInfo(ctx, msg.To[i])
			if err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}
		}

		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		html, err := s.replaceEmbeddedPictures(string(body), msg.Attachments, num, mailbox)
		if err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		out = &entity.MessageWithBody{
			Message: entity.Message{
				Id:        msg.SeqNum,
				To:        msg.To,
				From:      msg.From,
				SentDate:  msg.SentDate,
				Subject:   msg.Subject,
				IsRead:    msg.IsRead,
				Encrypted: msg.VaultId != "",
			},
			Body:        html,
			Attachments: msg.Attachments,
		}

		return nil
	}); err != nil {
		log.Error("cannot fetch message")
		return nil, err
	}

	return out, nil
}

func (s *Service) replaceEmbeddedPictures(body string, attachments []entity.Attachment, num uint32, mailbox string) (string, error) {

	for _, attachment := range attachments {

		cid := strings.Trim(attachment.ContentId, "<>")

		re, err := regexp.Compile(`cid:` + regexp.QuoteMeta(cid))
		if err != nil {
			slog.Debug("failed to compile regexp:", sl.Err(err))
		}

		body = re.ReplaceAllString(body, fmt.Sprintf(
			"%s/attachment/%s?mailnum=%d&mailbox=%s",
			s.cfg.Host(),
			cid,
			num,
			mailbox,
		))

	}

	return body, nil
}
