package port

import (
	"context"
	"crypt-link/core/domain"
)

// MessageService defines the inbound port for message business logic.
type MessageService interface {
	Create(ctx context.Context, id, encrypted string, expiresInSeconds int64) error
	GetAndInvalidate(ctx context.Context, id string) (*domain.Message, error)
}
