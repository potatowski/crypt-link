package service

import (
	"crypt-link/database"
	"crypt-link/model"
	"crypt-link/repository"
	"crypt-link/util"
	"errors"
	"time"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) Create(id, encrypted string, expiresAt time.Time) error {
	if expiresAt.Before(util.NowUTC()) {
		return errors.New("expiration time must be in the future")
	}

	if id == "" || encrypted == "" {
		return errors.New("id and encrypted content must not be empty")
	}

	repo := repository.NewMessageRepository(database.Client)
	msg := model.Message{
		ID:        id,
		Encrypted: encrypted,
		CreatedAt: util.NowUTC(),
		ExpiresAt: expiresAt,
	}
	return repo.Save(msg)
}

func (s *MessageService) GetAndInvalidate(id string) (*model.Message, error) {
	repo := repository.NewMessageRepository(database.Client)
	msg, err := repo.FindAndDelete(id)
	if err != nil {
		return nil, err
	}
	if util.NowUTC().After(msg.ExpiresAt) {
		return nil, errors.New("expired")
	}
	return msg, nil
}
