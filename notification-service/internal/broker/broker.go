package broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    string
}

func New(url, exchange, queue, bindingKey string) (*Broker, error) {
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
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err = ch.QueueBind(q.Name, bindingKey, exchange, false, nil); err != nil {
		conn.Close()
		return nil, err
	}
	return &Broker{conn: conn, ch: ch, q: q.Name}, nil
}

func (b *Broker) Consume() (<-chan amqp.Delivery, error) {
	return b.ch.Consume(b.q, "", false, false, false, false, nil)
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
