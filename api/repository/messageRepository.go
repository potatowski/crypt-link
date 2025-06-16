package repository

import (
	"context"
	"crypt-link/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository struct {
	Collection *mongo.Collection
}

func NewMessageRepository(client *mongo.Client) *MessageRepository {
	col := client.Database("onetime").Collection("messages")
	return &MessageRepository{Collection: col}
}

func (r *MessageRepository) Save(msg model.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.Collection.InsertOne(ctx, msg)
	return err
}

func (r *MessageRepository) FindAndDelete(id string) (*model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg model.Message
	err := r.Collection.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&msg)
	return &msg, err
}
