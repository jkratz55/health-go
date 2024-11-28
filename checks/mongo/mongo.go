package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// New returns a CheckFunc that verifies the connection to MongoDB by pinging it.
func New(client *mongo.Client, rp *readpref.ReadPref) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := client.Ping(ctx, rp)
		if err != nil {
			return fmt.Errorf("mongo ping: %w", err)
		}
		return nil
	}
}
