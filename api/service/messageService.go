package service

import (
	"crypt-link/model"
	"crypt-link/repository"
	"crypt-link/util"
	"errors"
	"time"
)

type MessageService struct {
	Repo *repository.MessageRepository
}

func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{Repo: repo}
}

func (s *MessageService) Create(id, encrypted string, expiresAt time.Time) error {
	msg := model.Message{
		ID:        id,
		Encrypted: encrypted,
		CreatedAt: util.NowUTC(),
		ExpiresAt: expiresAt,
	}
	return s.Repo.Save(msg)
}

func (s *MessageService) GetAndInvalidate(id string) (*model.Message, error) {
	msg, err := s.Repo.FindAndDelete(id)
	if err != nil {
		return nil, err
	}
	if util.NowUTC().After(msg.ExpiresAt) {
		return nil, errors.New("expired")
	}
	return msg, nil
}
