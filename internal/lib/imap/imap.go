package imap

import (
	"fmt"
	"log/slog"

	"github.com/emersion/go-imap/client"
	"github.com/tehrelt/unreal/internal/lib/logger/sl"
)

func Dial(email, password, host string, port int) (*client.Client, func(), error) {

	addr := fmt.Sprintf("%s:%d", host, port)

	slog.Debug("dialing imap", slog.String("addr", addr))
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		slog.Error("failed to dial at imap", sl.Err(err))
		return nil, nil, err
	}

	slog.Debug(
		"logging in imap",
		slog.String("email", email),
		slog.String("password", password),
	)

	if err := c.Login(email, password); err != nil {
		slog.Error("failed to login at imap", sl.Err(err))
		c.Close()
		return nil, nil, err
	}

	return c, func() {
		c.Logout()
		c.Close()
	}, nil
}
