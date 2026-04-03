package port

import (
	"crypt-link/core/domain"
	"time"
)

// MessageService defines the inbound port for message business logic.
type MessageService interface {
	Create(id, encrypted string, expiresAt time.Time) error
	GetAndInvalidate(id string) (*domain.Message, error)
}
