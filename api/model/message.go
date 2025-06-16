package model

import "time"

type Message struct {
	ID        string    `bson:"_id"`
	Encrypted string    `bson:"encrypted"`
	CreatedAt time.Time `bson:"createdAt"`
	ExpiresAt time.Time `bson:"expiresAt"`
}
