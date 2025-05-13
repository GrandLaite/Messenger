package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/broker"
	"notification-service/internal/notifier"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// ── параметры подключения ──────────────────────────────────────────────────
	rmqURL := getenv("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/")
	exchange := getenv("RABBIT_EXCHANGE", "msg.events")
	queue := getenv("RABBIT_QUEUE", "msg.notify")

	smtpHost := getenv("SMTP_HOST", "smtp.yandex.ru")
	smtpPort := getenv("SMTP_PORT", "587")
	smtpUser := getenv("SMTP_USERNAME", "")
	smtpPass := getenv("SMTP_PASSWORD", "")
	smtpFrom := getenv("SMTP_FROM", smtpUser)
	// ────────────────────────────────────────────────────────────────────────────

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	br, err := broker.New(rmqURL, exchange, queue, "message.created.*")
	if err != nil {
		logger.Error("rabbitmq dial", "err", err)
		os.Exit(1)
	}
	defer br.Close()

	mailer := notifier.NewEmailNotifier(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)

	deliveries, err := br.Consume()
	if err != nil {
		logger.Error("consume", "err", err)
		os.Exit(1)
	}

	// ловим Ctrl-C / docker stop
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		//----------------------------------------------------------------------
		case d := <-deliveries:
			var evt broker.MessageCreatedEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				logger.Warn("json unmarshal", "err", err)
				d.Nack(false, false) // отбросить
				continue
			}

			if evt.RecipientEmail == "" {
				logger.Warn("empty recipient email",
					"recipient", evt.Recipient, "msgID", evt.ID)
				d.Ack(false)
				continue
			}

			// ── новая тема и шаблон письма ───────────────────────────────────
			subject := "У Вас новое сообщение!"

			body := fmt.Sprintf(
				"Здравствуйте, %s!\n\n"+
					"У Вас новое сообщение от %s.\n\n"+
					"Текст сообщения:\n%s\n\n"+
					"-----------\n"+
					"Это автоматическое уведомление, отвечать на него не нужно.",
				evt.Recipient, // обращение
				evt.Sender,    // автор
				evt.Content,   // текст сообщения
			)
			// ─────────────────────────────────────────────────────────────────

			if err := mailer.Send(evt.RecipientEmail, subject, body); err != nil {
				logger.Error("email send",
					"to", evt.RecipientEmail, "err", err)
				d.Nack(false, true) // повторить позже
			} else {
				logger.Info("email sent",
					"to", evt.RecipientEmail, "msgID", evt.ID)
				d.Ack(false)
			}

		//----------------------------------------------------------------------
		case <-stop:
			logger.Info("notification-service shutting down")
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			<-ctx.Done()
			return
		}
	}
}
