package mongodb

import (
	"crypt-link/core/domain"
	"time"
)

type MessageEntity struct {
	ID        string    `bson:"_id"`
	Encrypted string    `bson:"encrypted"`
	CreatedAt time.Time `bson:"createdAt"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

func fromDomain(msg domain.Message) MessageEntity {
	return MessageEntity{
		ID:        msg.ID,
		Encrypted: msg.Encrypted,
		CreatedAt: msg.CreatedAt,
		ExpiresAt: msg.ExpiresAt,
	}
}

func (e MessageEntity) toDomain() *domain.Message {
	return &domain.Message{
		ID:        e.ID,
		Encrypted: e.Encrypted,
		CreatedAt: e.CreatedAt,
		ExpiresAt: e.ExpiresAt,
	}
}
