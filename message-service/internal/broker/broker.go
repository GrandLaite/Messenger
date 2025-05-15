package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
}

func New(url, exchange string) (*Broker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err = ch.Confirm(false); err != nil {
		conn.Close()
		return nil, err
	}
	return &Broker{conn: conn, ch: ch, exchange: exchange}, nil
}

func (b *Broker) PublishMessageCreated(evt MessageCreatedEvent) error {
	body, _ := json.Marshal(evt)
	rk := fmt.Sprintf("message.created.%s", evt.Recipient)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return b.ch.PublishWithContext(ctx, b.exchange, rk, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (b *Broker) Close() {
	if b.ch != nil {
		b.ch.Close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
}

type MessageCreatedEvent struct {
	ID             int    `json:"id"`
	Sender         string `json:"sender"`
	Recipient      string `json:"recipient"`
	RecipientEmail string `json:"recipient_email"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}
