package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	adapter_codec "github.com/vafrcor/go-cache/codec"
)

type GoRedisRepository struct {
	RDB    *redis.Client
	codec  adapter_codec.Codec
	Logger *log.Logger
}

func NewGoRedisRepository(client *redis.Client, codec adapter_codec.Codec) ICacheRepository {
	if codec == nil {
		codec = adapter_codec.JSONCodec{}
	}

	return &GoRedisRepository{
		RDB:    client,
		codec:  codec,
		Logger: log.Default(),
	}
}

func (c *GoRedisRepository) Set(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	b, err := c.codec.Marshal(value)
	if err != nil {
		c.Logger.Printf("GoRedisRepository.Set: %v", err)
		return err
	}

	if err := c.RDB.Set(ctx, key, b, ttl).Err(); err != nil {
		c.Logger.Printf("GoRedisRepository.Set: %v", err)
		return err
	}
	return nil
}

func (c *GoRedisRepository) Get(
	ctx context.Context,
	key string,
	dst any,
) error {
	b, err := c.RDB.Get(ctx, key).Bytes()
	if err != nil {
		c.Logger.Printf("GoRedisRepository.Get: %v", err)
		if err == redis.Nil {
			return ErrCacheNotFound
		}
		return err
	}

	return c.codec.Unmarshal(b, dst)
}

func (c *GoRedisRepository) Delete(ctx context.Context, key string) error {
	affected, err := c.RDB.Del(ctx, key).Result()
	if err != nil {
		c.Logger.Printf("GoRedisRepository.Delete: %v", err)
		return err
	}
	if affected == 0 {
		c.Logger.Printf("GoRedisRepository.Delete: cache not found (%s)", key)
	}
	return nil
}

func (c *GoRedisRepository) Exists(ctx context.Context, key string) bool {
	affected, err := c.RDB.Exists(ctx, key).Result()
	if err != nil {
		c.Logger.Printf("GoRedisRepository.Exists: %v", err)
		return false
	}
	if affected == 0 {
		c.Logger.Printf("GoRedisRepository.Exists: %v", ErrCacheNotFound)
	}
	return affected > 0
}

func (c *GoRedisRepository) Clear(ctx context.Context) error {
	if err := c.RDB.FlushDB(ctx).Err(); err != nil {
		c.Logger.Printf("GoRedisRepository.Clear: %v", err)
		return err
	}
	return nil
}

func (c *GoRedisRepository) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.RDB.TTL(ctx, key).Result()
	if err != nil {
		c.Logger.Printf("GoRedisRepository.TTL: %v", err)
		return 0, err
	}
	if ttl == -2*time.Second || ttl == -2*time.Nanosecond {
		c.Logger.Printf("GoRedisRepository.TTL: %v", ErrCacheNotFound)
		return 0, ErrCacheNotFound
	}
	if ttl == -time.Second || ttl == -time.Nanosecond {
		return 0, nil
	}
	return ttl, nil
}

func (c *GoRedisRepository) Touch(
	ctx context.Context,
	key string,
	ttl time.Duration,
) error {
	affected, err := c.RDB.Expire(ctx, key, ttl).Result()
	if err != nil {
		c.Logger.Printf("GoRedisRepository.Touch: %v", err)
		return err
	}
	if !affected {
		c.Logger.Printf("GoRedisRepository.Touch: %v", ErrCacheNotFound)
		return ErrCacheNotFound
	}
	return nil
}
