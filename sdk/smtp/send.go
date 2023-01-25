package smtp

import (
	"strings"

	"github.com/go-mail/mail"
	"golang.org/x/net/html"
)

//nolint:gochecknoglobals
var cfg SMTPConfiguration

type SMTPConfiguration struct {
	Host         string
	Port         int
	AuthUsername string
	AuthPassword string
	Sender       string
}

func (c SMTPConfiguration) SendEmail(recipient, cc []string, subject, body, attachmentPath string) error {
	m := mail.NewMessage()
	m.SetHeader("From", c.Sender)
	m.SetHeader("To", strings.Join(recipient, ","))
	m.SetHeader("Subject", subject)

	if len(cc) > 0 {
		m.SetHeader("Cc", strings.Join(cc, ","))
	}

	if isHTML(body) {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	if attachmentPath != "" {
		m.Attach(attachmentPath)
	}

	d := mail.NewDialer(c.Host, c.Port, c.AuthUsername, c.AuthPassword)

	return d.DialAndSend(m)
}

func isHTML(s string) bool {
	_, err := html.Parse(strings.NewReader(s))

	return err == nil
}
