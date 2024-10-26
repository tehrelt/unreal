package handlers

import (
	"strings"

	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/tehrelt/unreal/internal/dto"
	"github.com/tehrelt/unreal/internal/services/mailservice"
)

func SendMail(ms *mailservice.Service) echo.HandlerFunc {

	return func(c echo.Context) error {

		ctx, u, err := extractUser(c)
		if err != nil {
			return err
		}

		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(500, map[string]any{
				"error": err.Error(),
			})
		}

		req := &dto.SendMessageDto{
			From: &dto.MailRecord{Email: u.Email},
			To: lo.Map(form.Value["to"], func(email string, _ int) *dto.MailRecord {
				return &dto.MailRecord{Email: email}
			}),
			Attachments: form.File["attachment"],
		}

		if len(form.Value["subject"]) == 0 {
			return echo.NewHTTPError(400, "subject is required")
		}

		if len(form.Value["body"]) == 0 {
			return echo.NewHTTPError(400, "body is required")
		}

		req.Subject = form.Value["subject"][0]
		req.Body = strings.NewReader(form.Value["body"][0])

		if len(form.Value["encrypt"]) > 0 {
			val := form.Value["encrypt"][0]

			slog.Debug("encrypt mode", slog.String("val", val))

			if strings.Compare(val, "true") == 0 {
				req.DoEncryiption = true
			} else if strings.Compare(val, "false") == 0 {
				req.DoEncryiption = false
			} else {
				return echo.NewHTTPError(400, "incorrect encrypt value")
			}
		}

		if err := ms.Send(ctx, req); err != nil {
			return echo.NewHTTPError(500, err.Error())
		}

		return nil
	}
}
