ratelimit-bench
===============

Token bucket vs sliding window rate limiting in Go, both backed by Redis, under bursty load from k6.

I wanted a straight answer to an argument we kept having at work: is the sliding window worth the extra Redis calls? So I
put both behind the same tiny HTTP server and hit them with the same traffic.

Result on one node (8 vCPU, Redis on the same box), bursty profile in `k6/bursty.js`:

    limiter          req/s    p50     p99
    token bucket     15k      1.9ms   7.8ms
    sliding window   15k      2.1ms   5.6ms

The sliding window is slightly slower on median and noticeably steadier at p99, because the token bucket lets a full
bucket through at the start of every burst and Redis queues up behind it.

Run it
------

    docker run -d -p 6379:6379 redis:7
    go run ./cmd/server -limiter window -limit 100 -window 1s
    k6 run k6/bursty.js

Both limiters are single Lua scripts, so each check is one round trip and atomic.

Not done: a cluster setup, and a fixed window limiter (it was obviously worse at the edges, not worth benchmarking).
