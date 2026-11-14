package database

import (
	"context"
	"crypt-link/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const connectTimeout = 10 * time.Second

// Connect creates a MongoDB client and pings the primary so startup fails
// fast when the database is unreachable (mongo.Connect itself is lazy).
func Connect(ctx context.Context) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseUrl))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}
