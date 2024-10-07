package smtp

import (
	"fmt"
	"log/slog"
	"net/smtp"
)

func Dial(email, password, host string, port int) (smtp.Auth, error) {

	slog.Debug("dial smtp", slog.String("host", host), slog.Int("port", port))
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", email, password, host)

	slog.Debug("auth smtp", slog.String("email", email))
	if err := client.Auth(auth); err != nil {
		return nil, err
	}

	return auth, nil
}
