package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"ratelimit-bench/limiter"
	"github.com/redis/go-redis/v9"
)

func main() {
	kind := flag.String("limiter", "window", "bucket or window")
	limit := flag.Int("limit", 100, "requests allowed per window (bucket capacity)")
	window := flag.Duration("window", time.Second, "window length (bucket refill period)")
	addr := flag.String("addr", ":8080", "listen address")
	redisAddr := flag.String("redis", "localhost:6379", "redis address")
	flag.Parse()

	rdb := redis.NewClient(&redis.Options{Addr: *redisAddr, PoolSize: 64})

	var lim limiter.Limiter
	switch *kind {
	case "bucket":
		lim = limiter.NewTokenBucket(rdb, *limit, *window)
	case "window":
		lim = limiter.NewSlidingWindow(rdb, *limit, *window)
	default:
		log.Fatalf("unknown limiter %q", *kind)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Client")
		if key == "" {
			key = "anon"
		}
		ok, err := lim.Allow(r.Context(), key)
		if err != nil {
			http.Error(w, "limiter error", http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok\n"))
	})

	log.Printf("%s limiter on %s", *kind, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
