package service

import (
	"errors"
	"log/slog"

	"message-service/internal/broker"
	"message-service/internal/repository"
)

type MessageService struct {
	repo   *repository.MessageRepository
	broker *broker.Broker
}

func NewMessageService(r *repository.MessageRepository, b *broker.Broker) *MessageService {
	return &MessageService{repo: r, broker: b}
}

func (s *MessageService) Create(sender, recipient, content string) (repository.Message, error) {
	msg, err := s.repo.Create(sender, recipient, content)
	if err != nil {
		return msg, err
	}

	if s.broker != nil {
		email, emErr := getRecipientEmail(recipient)
		if emErr != nil {
			slog.Default().Warn("user email lookup failed", "err", emErr)
		}

		_ = s.broker.PublishMessageCreated(broker.MessageCreatedEvent{
			ID:             msg.ID,
			Sender:         msg.SenderNickname,
			Recipient:      msg.RecipientNickname,
			RecipientEmail: email,
			Content:        msg.Content,
			CreatedAt:      msg.CreatedAt,
		})
	}
	return msg, nil
}

func (s *MessageService) GetByID(id int) (repository.Message, error) {
	return s.repo.GetByID(id)
}

func (s *MessageService) Delete(id int, requester string) error {
	msg, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if msg.SenderNickname != requester {
		return errors.New("forbidden")
	}
	return s.repo.Delete(id)
}

func (s *MessageService) GetConversation(u1, u2 string) ([]repository.Message, error) {
	return s.repo.GetConversation(u1, u2)
}

func (s *MessageService) GetDialogs(nickname string) ([]string, error) {
	return s.repo.GetDialogs(nickname)
}
