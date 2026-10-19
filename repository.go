package cache

import (
	"context"
	"time"
)

type ICacheRepository interface {
	Get(ctx context.Context, key string, destination any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
	Clear(ctx context.Context) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Touch(ctx context.Context, key string, ttl time.Duration) error
}
