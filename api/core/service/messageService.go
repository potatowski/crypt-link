package service

import (
	"context"
	"crypt-link/core/domain"
	"crypt-link/core/port"
	"fmt"
	"time"
)

type messageService struct {
	repo port.MessageRepository
	now  func() time.Time
}

// NewMessageService creates a MessageService. now is the clock used for
// timestamps and expiry checks; if nil, the current UTC time is used.
func NewMessageService(repo port.MessageRepository, now func() time.Time) port.MessageService {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &messageService{
		repo: repo,
		now:  now,
	}
}

func (s *messageService) Create(ctx context.Context, id, encrypted string, expiresAt time.Time) error {
	now := s.now()
	if expiresAt.Before(now) {
		return fmt.Errorf("%w: expiration time must be in the future", domain.ErrInvalidInput)
	}

	if id == "" || encrypted == "" {
		return fmt.Errorf("%w: id and encrypted content must not be empty", domain.ErrInvalidInput)
	}

	msg := domain.Message{
		ID:        id,
		Encrypted: encrypted,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	return s.repo.Save(ctx, msg)
}

func (s *messageService) GetAndInvalidate(ctx context.Context, id string) (*domain.Message, error) {
	msg, err := s.repo.FindAndDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if msg.IsExpired(s.now()) {
		return nil, domain.ErrExpired
	}
	return msg, nil
}
