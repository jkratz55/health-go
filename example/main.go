package main

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/jkratz55/health-go"
)

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	hc := health.New()
	hc.Register(health.Component{
		Name:     "redis",
		Critical: true,
		Check: func(ctx context.Context) health.Status {
			res, err := rdb.Ping(ctx).Result()
			if err != nil {
				return health.StatusDown
			}
			if res != "PONG" {
				return health.StatusDown
			}
			return health.StatusUp
		},
	})

	if err := health.EnablePrometheus(hc); err != nil {
		panic(err)
	}

	go func() {
		promServer := http.Server{
			Addr:    ":8082",
			Handler: promhttp.Handler(),
		}
		promServer.ListenAndServe()
	}()

	http.HandleFunc("/health", hc.HandlerFunc())
	http.ListenAndServe(":8080", nil)
}
