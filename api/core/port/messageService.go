package port

import (
	"context"
	"crypt-link/core/domain"
	"time"
)

// MessageService defines the inbound port for message business logic.
type MessageService interface {
	Create(ctx context.Context, id, encrypted string, expiresAt time.Time) error
	GetAndInvalidate(ctx context.Context, id string) (*domain.Message, error)
}
