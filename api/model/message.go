package model

import "time"

// Message represents an encrypted message with metadata including its unique identifier,
// encrypted content, creation time, and expiration time. It is mapped to a BSON document
// for storage in a MongoDB collection.
type Message struct {
	ID        string    `bson:"_id"`
	Encrypted string    `bson:"encrypted"`
	CreatedAt time.Time `bson:"createdAt"`
	ExpiresAt time.Time `bson:"expiresAt"`
}
