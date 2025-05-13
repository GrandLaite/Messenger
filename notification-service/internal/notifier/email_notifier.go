package notifier

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
)

// EmailNotifier отправляет письма через SMTP-сервер
type EmailNotifier struct {
	host string
	port string
	user string
	pass string
	from string
}

// NewEmailNotifier создаёт почтовый нотификатор
func NewEmailNotifier(host, port, user, pass, from string) *EmailNotifier {
	return &EmailNotifier{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

// Send отправляет письмо получателю to с темой subject и телом body
func (e *EmailNotifier) Send(to, subject, body string) error {
	// --- корректно кодируем тему, если есть не-ASCII символы -------------
	// RFC 2047 «encoded-word» (Q-encoding)
	encodedSubj := mime.QEncoding.Encode("utf-8", subject)

	// собираем MIME-сообщение
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

	// --- устанавливаем соединение и шифрование --------------------------
	conn, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// STARTTLS, если поддерживается
	if ok, _ := conn.Extension("STARTTLS"); ok {
		cfg := &tls.Config{ServerName: e.host}
		if err = conn.StartTLS(cfg); err != nil {
			return err
		}
	}

	// аутентификация, заголовки SMTP и передача тела письма
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
