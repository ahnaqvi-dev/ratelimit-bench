package limiter

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// Sliding log in a sorted set. KEYS[1] key; ARGV: limit, window ms, now ms, member
var windowScript = redis.NewScript(`
local limit = tonumber(ARGV[1])
local win = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
redis.call("ZREMRANGEBYSCORE", KEYS[1], 0, now - win)
if redis.call("ZCARD", KEYS[1]) < limit then
  redis.call("ZADD", KEYS[1], now, ARGV[4])
  redis.call("PEXPIRE", KEYS[1], win)
  return 1
end
return 0
`)

type SlidingWindow struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
	seq    atomic.Uint64
}

func NewSlidingWindow(rdb *redis.Client, limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{rdb: rdb, limit: limit, window: window}
}

func (s *SlidingWindow) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	member := strconv.FormatInt(now, 10) + "-" + strconv.FormatUint(s.seq.Add(1), 10)
	n, err := windowScript.Run(ctx, s.rdb, []string{"sw:" + key}, s.limit, ms(s.window), now, member).Int()
	return n == 1, err
}
