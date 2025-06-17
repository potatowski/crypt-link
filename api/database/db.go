package database

import (
	"context"
	"crypt-link/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var Ctx = context.Background()

func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.DatabaseUrl))

	if err != nil {
		return err
	}

	Client = client
	Ctx = ctx
	return err
}
