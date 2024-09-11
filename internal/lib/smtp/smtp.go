package smtp

import (
	"fmt"
	"net/smtp"
)

func Dial(email, password, host string, port int) (*smtp.Client, func(), error) {
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, nil, err
	}

	if err := client.Auth(smtp.PlainAuth("", email, password, host)); err != nil {
		return nil, nil, err
	}

	return client, func() {
		client.Close()
	}, nil
}
