package cache

import (
	"context"
	"log"
	"time"

	"github.com/coocood/freecache"
	adapter_codec "github.com/vafrcor/go-cache/codec"
)

type FreecacheRepository struct {
	fc     *freecache.Cache
	codec  adapter_codec.Codec
	Logger *log.Logger
}

func NewFreecacheRepository(size int, codec adapter_codec.Codec) ICacheRepository {
	if codec == nil {
		codec = adapter_codec.JSONCodec{}
	}

	return &FreecacheRepository{
		fc:     freecache.NewCache(size),
		codec:  codec,
		Logger: log.Default(),
	}
}

func (c *FreecacheRepository) Set(
	_ context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	b, err := c.codec.Marshal(value)
	if err != nil {
		c.Logger.Printf("FreecacheRepository.Set: %v", err)
		return err
	}

	return c.fc.Set(
		[]byte(key),
		b,
		int(ttl.Seconds()),
	)
}

func (c *FreecacheRepository) Get(
	_ context.Context,
	key string,
	dst any,
) error {
	b, err := c.fc.Get([]byte(key))
	if err != nil {
		c.Logger.Printf("FreecacheRepository.Get: %v", err)
		return ErrCacheNotFound
	}

	return c.codec.Unmarshal(b, dst)
}

func (c *FreecacheRepository) Delete(_ context.Context, key string) error {
	affected := c.fc.Del([]byte(key))
	if !affected {
		c.Logger.Printf("FreecacheRepository.Delete: cache not found (%s)", key)
	}
	return nil
}

func (c *FreecacheRepository) Exists(_ context.Context, key string) bool {
	_, err := c.fc.Get([]byte(key))
	if err != nil {
		c.Logger.Printf("FreecacheRepository.Exists: %v", err)
	}
	return err == nil
}

func (c *FreecacheRepository) Clear(_ context.Context) error {
	c.fc.Clear()
	return nil
}

func (c *FreecacheRepository) TTL(_ context.Context, key string) (time.Duration, error) {
	seconds, err := c.fc.TTL([]byte(key))
	if err != nil {
		c.Logger.Printf("FreecacheRepository.TTL: %v", err)
		return 0, ErrCacheNotFound
	}

	return time.Duration(seconds) * time.Second, nil
}

func (c *FreecacheRepository) Touch(
	_ context.Context,
	key string,
	ttl time.Duration,
) error {
	value, err := c.fc.Get([]byte(key))
	if err != nil {
		c.Logger.Printf("FreecacheRepository.Touch: %v", err)
		return ErrCacheNotFound
	}

	return c.fc.Set(
		[]byte(key),
		value,
		int(ttl.Seconds()),
	)
}
