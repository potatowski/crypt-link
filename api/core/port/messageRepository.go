package port

import (
	"context"
	"crypt-link/core/domain"
)

// MessageRepository defines the outbound port for message storage operations.
// FindAndDelete returns domain.ErrNotFound when no message matches the id.
type MessageRepository interface {
	Save(ctx context.Context, msg domain.Message) error
	FindAndDelete(ctx context.Context, id string) (*domain.Message, error)
}
