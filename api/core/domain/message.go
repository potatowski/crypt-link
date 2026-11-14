package domain

import (
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a message does not exist (or was already read).
	ErrNotFound = errors.New("message not found")
	// ErrExpired is returned when a message exists but its expiration time has passed.
	ErrExpired = errors.New("message expired")
	// ErrInvalidInput is returned when the provided message data is invalid.
	ErrInvalidInput = errors.New("invalid input")
)

// Message represents an encrypted message with metadata including its unique identifier,
// encrypted content, creation time, and expiration time.
type Message struct {
	ID        string
	Encrypted string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// IsExpired reports whether the message is expired at the given time.
func (m Message) IsExpired(now time.Time) bool {
	return now.After(m.ExpiresAt)
}
