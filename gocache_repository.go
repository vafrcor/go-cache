package cache

import (
	"context"
	"log"
	"time"

	gocache "github.com/patrickmn/go-cache"
	adapter_codec "github.com/vafrcor/go-cache/codec"
)

type GoCacheRepository struct {
	gc     *gocache.Cache
	codec  adapter_codec.Codec
	Logger *log.Logger
}

func NewGoCacheRepository(
	defaultExpiration time.Duration,
	cleanupInterval time.Duration,
	codec adapter_codec.Codec,
) ICacheRepository {
	if codec == nil {
		codec = adapter_codec.JSONCodec{}
	}

	return &GoCacheRepository{
		gc:     gocache.New(defaultExpiration, cleanupInterval),
		codec:  codec,
		Logger: log.Default(),
	}
}

func (c *GoCacheRepository) Set(
	_ context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	b, err := c.codec.Marshal(value)
	if err != nil {
		c.Logger.Printf("GoCacheRepository.Set: %v", err)
		return err
	}

	c.gc.Set(key, b, ttl)
	return nil
}

func (c *GoCacheRepository) Get(
	_ context.Context,
	key string,
	dst any,
) error {
	value, found := c.gc.Get(key)
	if !found {
		c.Logger.Printf("GoCacheRepository.Get: %v", ErrCacheNotFound)
		return ErrCacheNotFound
	}

	return c.codec.Unmarshal(value.([]byte), dst)
}

func (c *GoCacheRepository) Delete(_ context.Context, key string) error {
	if _, found := c.gc.Get(key); !found {
		c.Logger.Printf("GoCacheRepository.Delete: cache not found (%s)", key)
	}
	c.gc.Delete(key)
	return nil
}

func (c *GoCacheRepository) Exists(_ context.Context, key string) bool {
	_, found := c.gc.Get(key)
	if !found {
		c.Logger.Printf("GoCacheRepository.Exists: %v", ErrCacheNotFound)
	}
	return found
}

func (c *GoCacheRepository) Clear(_ context.Context) error {
	c.gc.Flush()
	return nil
}

func (c *GoCacheRepository) TTL(_ context.Context, key string) (time.Duration, error) {
	_, expiration, found := c.gc.GetWithExpiration(key)
	if !found {
		c.Logger.Printf("GoCacheRepository.TTL: %v", ErrCacheNotFound)
		return 0, ErrCacheNotFound
	}
	if expiration.IsZero() {
		return 0, nil
	}

	return time.Until(expiration), nil
}

func (c *GoCacheRepository) Touch(
	_ context.Context,
	key string,
	ttl time.Duration,
) error {
	value, found := c.gc.Get(key)
	if !found {
		c.Logger.Printf("GoCacheRepository.Touch: %v", ErrCacheNotFound)
		return ErrCacheNotFound
	}

	c.gc.Set(key, value, ttl)
	return nil
}
