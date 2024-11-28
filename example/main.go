package main

import (
	"context"
	"errors"
	"net/http"
	"time"

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
		Timeout:  time.Second * 1,
		Interval: time.Second * 5,
		Check: func(ctx context.Context) error {
			res, err := rdb.Ping(ctx).Result()
			if err != nil {
				return err
			}
			if res != "PONG" {
				return errors.New("unexpected response")
			}
			return nil
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
