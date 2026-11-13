package limiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// KEYS[1] bucket key; ARGV: capacity, refill per ms, now ms
var bucketScript = redis.NewScript(`
local cap = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local b = redis.call("HMGET", KEYS[1], "tokens", "ts")
local tokens = tonumber(b[1]) or cap
local ts = tonumber(b[2]) or now
tokens = math.min(cap, tokens + (now - ts) * rate)
local ok = 0
if tokens >= 1 then
  tokens = tokens - 1
  ok = 1
end
redis.call("HSET", KEYS[1], "tokens", tokens, "ts", now)
redis.call("PEXPIRE", KEYS[1], math.ceil(cap / rate) + 1000)
return ok
`)

type TokenBucket struct {
	rdb      *redis.Client
	capacity int
	per      time.Duration
}

func NewTokenBucket(rdb *redis.Client, capacity int, per time.Duration) *TokenBucket {
	return &TokenBucket{rdb: rdb, capacity: capacity, per: per}
}

func (t *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	rate := float64(t.capacity) / float64(ms(t.per))
	n, err := bucketScript.Run(ctx, t.rdb, []string{"tb:" + key}, t.capacity, rate, time.Now().UnixMilli()).Int()
	return n == 1, err
}
