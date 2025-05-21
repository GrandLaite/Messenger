package email

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"os"
	"strings"

	"notification-service/internal/notifier"
)

// EmailNotifier удовлетворяет интерфейсу notifier.Notifier.
type EmailNotifier struct {
	host, port, user, pass, from string
}

// NewFromEnv читает переменные окружения и возвращает готовый объект.
func NewFromEnv() (notifier.Notifier, error) {
	get := func(k, d string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return d
	}
	return &EmailNotifier{
		host: get("SMTP_HOST", "smtp.yandex.ru"),
		port: get("SMTP_PORT", "587"),
		user: get("SMTP_USERNAME", ""),
		pass: get("SMTP_PASSWORD", ""),
		from: get("SMTP_FROM", get("SMTP_USERNAME", "")),
	}, nil
}

func init() {
	// «Плагин» регистрируется под именем "email"
	notifier.Register("email", NewFromEnv)
}

// ---------------- реализация интерфейса ----------------

func (e *EmailNotifier) Send(to, subject, body string) error {
	encSubj := mime.QEncoding.Encode("utf-8", subject)
	msg := strings.Join([]string{
		"From: " + e.from,
		"To: " + to,
		"Subject: " + encSubj,
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
		if err = conn.StartTLS(&tls.Config{ServerName: e.host}); err != nil {
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
	w, err := conn.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = fmt.Fprint(w, msg)
	return err
}
