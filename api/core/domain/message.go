package domain

import "time"

// Message represents an encrypted message with metadata including its unique identifier,
// encrypted content, creation time, and expiration time.
type Message struct {
	ID        string
	Encrypted string
	CreatedAt time.Time
	ExpiresAt time.Time
}
