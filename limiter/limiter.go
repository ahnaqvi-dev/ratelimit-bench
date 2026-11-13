package limiter

import (
	"context"
	"time"
)

// Limiter decides whether one more request for key is allowed right now.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

func ms(d time.Duration) int64 { return d.Milliseconds() }
