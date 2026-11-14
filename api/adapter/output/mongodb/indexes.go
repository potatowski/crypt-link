package mongodb

import (
	"context"
	"crypt-link/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates the indexes required by the message collection.
// It is idempotent: creating an index that already exists with the same
// definition is a no-op in MongoDB.
//
// The TTL index on expiresAt lets MongoDB purge expired messages that were
// never read.
func EnsureIndexes(ctx context.Context, client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	col := client.Database(config.DatabaseName).Collection(collectionName)
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetName("expiresAt_ttl").SetExpireAfterSeconds(0),
	})
	return err
}
