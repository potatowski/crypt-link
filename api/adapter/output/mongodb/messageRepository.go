package mongodb

import (
	"context"
	"crypt-link/config"
	"crypt-link/core/domain"
	"crypt-link/core/port"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const opTimeout = 5 * time.Second

type messageRepository struct {
	Collection *mongo.Collection
}

func NewMessageRepository(client *mongo.Client) port.MessageRepository {
	col := client.Database(config.DatabaseName).Collection("messages")
	return &messageRepository{Collection: col}
}

func (r *messageRepository) Save(ctx context.Context, msg domain.Message) error {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	entity := fromDomain(msg)
	_, err := r.Collection.InsertOne(ctx, entity)
	return err
}

func (r *messageRepository) FindAndDelete(ctx context.Context, id string) (*domain.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()

	var entity MessageEntity
	err := r.Collection.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&entity)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return entity.toDomain(), nil
}
