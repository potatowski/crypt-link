package port

import "crypt-link/core/domain"

// MessageRepository defines the outbound port for message storage operations.
type MessageRepository interface {
	Save(msg domain.Message) error
	FindAndDelete(id string) (*domain.Message, error)
}
