package checks

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func Redis(client redis.UniversalClient, timeout time.Duration) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		_, err := client.Ping(ctx).Result()
		return err
	}
}
