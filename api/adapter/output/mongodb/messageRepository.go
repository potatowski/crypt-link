package mongodb

import (
	"context"
	"crypt-link/core/domain"
	"crypt-link/core/port"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type messageRepository struct {
	Collection *mongo.Collection
}

func NewMessageRepository(client *mongo.Client) port.MessageRepository {
	col := client.Database("onetime").Collection("messages")
	return &messageRepository{Collection: col}
}

func (r *messageRepository) Save(msg domain.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.Collection.InsertOne(ctx, msg)
	return err
}

func (r *messageRepository) FindAndDelete(id string) (*domain.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg domain.Message
	err := r.Collection.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
