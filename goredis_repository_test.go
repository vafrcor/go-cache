package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type goRedisTestPayload struct {
	Name string `json:"name"`
}

func newTestGoRedisRepository(t *testing.T) *GoRedisRepository {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	repo, ok := NewGoRedisRepository(client, nil).(*GoRedisRepository)
	if !ok {
		t.Fatal("NewGoRedisRepository() did not return *GoRedisRepository")
	}

	return repo
}

func TestNewGoRedisRepositoryReturnsCacheRepository(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	var _ ICacheRepository = NewGoRedisRepository(client, nil)
}

func TestGoRedisRepositorySetGetAndExists(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	err := repo.Set(ctx, "payload", goRedisTestPayload{Name: "user"}, time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !repo.Exists(ctx, "payload") {
		t.Fatal("Exists() = false, want true")
	}

	var got goRedisTestPayload
	err = repo.Get(ctx, "payload", &got)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != "user" {
		t.Fatalf("Get() = %+v, want name user", got)
	}
}

func TestGoRedisRepositorySetReturnsMarshalError(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	err := repo.Set(ctx, "invalid", make(chan int), time.Minute)
	if err == nil {
		t.Fatal("Set() error = nil, want marshal error")
	}
}

func TestGoRedisRepositoryGetMissingReturnsErrCacheNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	var got goRedisTestPayload
	err := repo.Get(ctx, "missing", &got)
	if !errors.Is(err, ErrCacheNotFound) {
		t.Fatalf("Get() error = %v, want ErrCacheNotFound", err)
	}
}

func TestGoRedisRepositoryDelete(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	if err := repo.Set(ctx, "delete", "value", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := repo.Delete(ctx, "delete"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var got string
	err := repo.Get(ctx, "delete", &got)
	if !errors.Is(err, ErrCacheNotFound) {
		t.Fatalf("Get() after Delete() error = %v, want ErrCacheNotFound", err)
	}

	if err := repo.Delete(ctx, "missing"); err != nil {
		t.Fatalf("Delete() missing error = %v, want nil", err)
	}
}

func TestGoRedisRepositoryExistsReturnsFalseForMissingKey(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	if repo.Exists(ctx, "missing") {
		t.Fatal("Exists() = true, want false")
	}
}

func TestGoRedisRepositoryClear(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	if err := repo.Set(ctx, "first", "value", time.Minute); err != nil {
		t.Fatalf("Set(first) error = %v", err)
	}
	if err := repo.Set(ctx, "second", "value", time.Minute); err != nil {
		t.Fatalf("Set(second) error = %v", err)
	}
	if err := repo.Clear(ctx); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	if repo.Exists(ctx, "first") || repo.Exists(ctx, "second") {
		t.Fatal("Exists() after Clear() = true, want false")
	}
}

func TestGoRedisRepositoryTTL(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	if err := repo.Set(ctx, "ttl", "value", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	ttl, err := repo.TTL(ctx, "ttl")
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Fatalf("TTL() = %s, want within (0, 1m]", ttl)
	}

	if _, err := repo.TTL(ctx, "missing"); !errors.Is(err, ErrCacheNotFound) {
		t.Fatalf("TTL() missing error = %v, want ErrCacheNotFound", err)
	}
}

func TestGoRedisRepositoryTTLNoExpiration(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	if err := repo.Set(ctx, "ttl", "value", 0); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	ttl, err := repo.TTL(ctx, "ttl")
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if ttl != 0 {
		t.Fatalf("TTL() = %s, want 0 for no expiration", ttl)
	}
}

func TestGoRedisRepositoryTouch(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoRedisRepository(t)

	if err := repo.Set(ctx, "touch", "value", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := repo.Touch(ctx, "touch", 2*time.Minute); err != nil {
		t.Fatalf("Touch() error = %v", err)
	}

	ttl, err := repo.TTL(ctx, "touch")
	if err != nil {
		t.Fatalf("TTL() after Touch() error = %v", err)
	}
	if ttl <= time.Minute || ttl > 2*time.Minute {
		t.Fatalf("TTL() after Touch() = %s, want within (1m, 2m]", ttl)
	}

	var got string
	if err := repo.Get(ctx, "touch", &got); err != nil {
		t.Fatalf("Get() after Touch() error = %v", err)
	}
	if got != "value" {
		t.Fatalf("Get() after Touch() = %q, want value", got)
	}

	if err := repo.Touch(ctx, "missing", time.Minute); !errors.Is(err, ErrCacheNotFound) {
		t.Fatalf("Touch() missing error = %v, want ErrCacheNotFound", err)
	}
}
