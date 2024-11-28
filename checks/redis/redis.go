package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
}

// New returns a CheckFunc that verifies the connection to Redis by pinging it.
//
// Note that a successful ping simply means that the connection to the server is
// open and the server is responding. It does not guarantee that the client is
// authenticated or can perform operations.
func New(client redis.UniversalClient) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		res, err := client.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("redis ping failed: %w", err)
		}
		if res != "PONG" {
			return fmt.Errorf("unexpected response from redis: %s", res)
		}
		return err
	}
}

// VerifyReadWrites returns a CheckFunc that verifies the ability to perform
// SET, GET, and DEL operations on Redis.
//
// If you simply want to check the connection to Redis, use the New function as
// it is more efficient as it only performs a PING operation. This function is
// more thorough and should be used when you want to verify the ability to read
// and write data to Redis.
func VerifyReadWrites(client redis.UniversalClient) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := client.Set(ctx, hostname+":healthcheck", "hello", time.Minute*1).Err()
		if err != nil {
			return fmt.Errorf("redis SET: %w", err)
		}
		res, err := client.Get(ctx, hostname+":healthcheck").Result()
		if err != nil {
			return fmt.Errorf("redis GET: %w", err)
		}
		if res != "hello" {
			return fmt.Errorf("unexpected value for key: %s", hostname+":healthcheck")
		}
		err = client.Del(ctx, hostname+":healthcheck").Err()
		if err != nil {
			return fmt.Errorf("redis DEL: %w", err)
		}
		return nil
	}
}
