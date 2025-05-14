package notifier

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
)

type EmailNotifier struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewEmailNotifier(host, port, user, pass, from string) *EmailNotifier {
	return &EmailNotifier{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

func (e *EmailNotifier) Send(to, subject, body string) error {
	encodedSubj := mime.QEncoding.Encode("utf-8", subject)

	msg := strings.Join([]string{
		"From: " + e.from,
		"To: " + to,
		"Subject: " + encodedSubj,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		body,
	}, "\r\n")

	addr := net.JoinHostPort(e.host, e.port)
	auth := smtp.PlainAuth("", e.user, e.pass, e.host)

	conn, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if ok, _ := conn.Extension("STARTTLS"); ok {
		cfg := &tls.Config{ServerName: e.host}
		if err = conn.StartTLS(cfg); err != nil {
			return err
		}
	}

	if err = conn.Auth(auth); err != nil {
		return err
	}
	if err = conn.Mail(e.from); err != nil {
		return err
	}
	if err = conn.Rcpt(to); err != nil {
		return err
	}

	wc, err := conn.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	_, err = fmt.Fprint(wc, msg)
	return err
}
