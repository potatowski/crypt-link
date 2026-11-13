package service

import (
	"context"
	"crypt-link/core/domain"
	"crypt-link/core/port"
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

// Create stores a new message. expiresInSeconds selects the message lifetime;
// zero uses domain.DefaultExpiry.
func (s *messageService) Create(ctx context.Context, id, encrypted string, expiresInSeconds int64) error {
	expiresIn, err := domain.ExpiryFromSeconds(expiresInSeconds)
	if err != nil {
		return err
	}

	if err := domain.ValidateID(id); err != nil {
		return err
	}

	if err := domain.ValidateEncrypted(encrypted); err != nil {
		return err
	}

	now := s.now()
	msg := domain.Message{
		ID:        id,
		Encrypted: encrypted,
		CreatedAt: now,
		ExpiresAt: now.Add(expiresIn),
	}
	return s.repo.Save(ctx, msg)
}

func (s *messageService) GetAndInvalidate(ctx context.Context, id string) (*domain.Message, error) {
	if err := domain.ValidateID(id); err != nil {
		return nil, err
	}

	msg, err := s.repo.FindAndDelete(ctx, id)
	if err != nil {
		return nil, err
	}
	if msg.IsExpired(s.now()) {
		return nil, domain.ErrExpired
	}
	return msg, nil
}
