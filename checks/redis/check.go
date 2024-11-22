package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jkratz55/health-go"
)

type Pinger interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

func New(client Pinger, sla time.Duration) func(context.Context) health.Status {
	return func(ctx context.Context) health.Status {
		start := time.Now()
		res, err := client.Ping(ctx).Result()
		elapsed := time.Since(start)
		if err != nil {
			return health.StatusDown
		}
		if res != "PONG" {
			return health.StatusDown
		}

		if elapsed > sla {
			return health.StatusDegraded
		}

		return health.StatusUp
	}
}
