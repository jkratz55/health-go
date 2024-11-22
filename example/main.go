package main

func main() {

	// rdb := redis.NewClient(&redis.Options{
	// 	Addr: "localhost:6379",
	// })
	//
	// hc := health.New()
	// hc.Register(health.Component{
	// 	Name:     "redis",
	// 	Critical: true,
	// 	Check: func(ctx context.Context) bool {
	// 		res, err := rdb.Ping(ctx).Result()
	// 		if err != nil {
	// 			return false
	// 		}
	// 		if res != "PONG" {
	// 			return false
	// 		}
	// 		return true
	// 	},
	// })
	//
	// if err := health.EnablePrometheus(hc); err != nil {
	// 	panic(err)
	// }
	//
	// go func() {
	// 	promServer := http.Server{
	// 		Addr:    ":8082",
	// 		Handler: promhttp.Handler(),
	// 	}
	// 	promServer.ListenAndServe()
	// }()
	//
	// http.HandleFunc("/health", hc.HandlerFunc())
	// http.ListenAndServe(":8080", nil)
}
