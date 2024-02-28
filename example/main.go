package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/jkratz55/health-go"
	checkRedis "github.com/jkratz55/health-go/checks/redis"
)

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	hc := health.New()
	hc.Register(health.Component{
		Name:     "redis",
		Critical: false,
		// Check: health.CheckFunc(func(ctx context.Context) health.Status {
		// 	res, err := rdb.Ping(ctx).Result()
		// 	if err != nil {
		// 		return health.StatusDown
		// 	}
		// 	if res != "PONG" {
		// 		return health.StatusDown
		// 	}
		// 	return health.StatusUp
		// }),
		Check: health.CheckFunc(checkRedis.New(rdb, 0)),
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
