package rueidis

import (
	"context"
	"fmt"

	"github.com/redis/rueidis"
)

// New returns a CheckFunc that verifies the connection to Redis by pinging it.
//
// Note that a successful ping simply means that the connection to the server is
// open and the server is responding. It does not guarantee that the client is
// authenticated or can perform operations.
func New(client rueidis.Client) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		res, err := client.Do(ctx, client.B().Ping().Build()).ToString()
		if err != nil {
			return fmt.Errorf("redis ping: %w", err)
		}
		if res != "PONG" {
			return fmt.Errorf("redis: unexpected response to ping: %s", res)
		}
		return nil
	}
}
