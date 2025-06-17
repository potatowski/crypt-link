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

// Save inserts a new Message into the MongoDB collection.
// It creates a context with a 5-second timeout for the operation.
// Returns an error if the insertion fails.
func (r *MessageRepository) Save(msg model.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.Collection.InsertOne(ctx, msg)
	return err
}

// FindAndDelete searches for a message by its ID in the MongoDB collection,
// deletes it if found, and returns the deleted message. If the message is not found
// or an error occurs during the operation, it returns an error.
// The operation is performed with a timeout of 5 seconds.
func (r *MessageRepository) FindAndDelete(id string) (*model.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg model.Message
	err := r.Collection.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&msg)
	return &msg, err
}
