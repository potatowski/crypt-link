package service

import (
	"crypt-link/core/domain"
	"crypt-link/core/port"
	"crypt-link/util"
	"errors"
	"time"
)

type messageService struct {
	repo port.MessageRepository
}

func NewMessageService(repo port.MessageRepository) port.MessageService {
	return &messageService{
		repo: repo,
	}
}

func (s *messageService) Create(id, encrypted string, expiresAt time.Time) error {
	if expiresAt.Before(util.NowUTC()) {
		return errors.New("expiration time must be in the future")
	}

	if id == "" || encrypted == "" {
		return errors.New("id and encrypted content must not be empty")
	}

	msg := domain.Message{
		ID:        id,
		Encrypted: encrypted,
		CreatedAt: util.NowUTC(),
		ExpiresAt: expiresAt,
	}
	return s.repo.Save(msg)
}

func (s *messageService) GetAndInvalidate(id string) (*domain.Message, error) {
	msg, err := s.repo.FindAndDelete(id)
	if err != nil {
		return nil, err
	}
	if util.NowUTC().After(msg.ExpiresAt) {
		return nil, errors.New("expired")
	}
	return msg, nil
}
