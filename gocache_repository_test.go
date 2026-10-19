package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type goCacheTestPayload struct {
	Name string `json:"name"`
}

func newTestGoCacheRepository(t *testing.T) *GoCacheRepository {
	t.Helper()

	repo, ok := NewGoCacheRepository(time.Minute, time.Minute, nil).(*GoCacheRepository)
	if !ok {
		t.Fatal("NewGoCacheRepository() did not return *GoCacheRepository")
	}

	return repo
}

func TestNewGoCacheRepositoryReturnsCacheRepository(t *testing.T) {
	var _ ICacheRepository = NewGoCacheRepository(time.Minute, time.Minute, nil)
}

func TestGoCacheRepositorySetGetAndExists(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

	err := repo.Set(ctx, "payload", goCacheTestPayload{Name: "user"}, time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if !repo.Exists(ctx, "payload") {
		t.Fatal("Exists() = false, want true")
	}

	var got goCacheTestPayload
	err = repo.Get(ctx, "payload", &got)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != "user" {
		t.Fatalf("Get() = %+v, want name user", got)
	}
}

func TestGoCacheRepositorySetReturnsMarshalError(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

	err := repo.Set(ctx, "invalid", make(chan int), time.Minute)
	if err == nil {
		t.Fatal("Set() error = nil, want marshal error")
	}
}

func TestGoCacheRepositoryGetMissingReturnsErrCacheNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

	var got goCacheTestPayload
	err := repo.Get(ctx, "missing", &got)
	if !errors.Is(err, ErrCacheNotFound) {
		t.Fatalf("Get() error = %v, want ErrCacheNotFound", err)
	}
}

func TestGoCacheRepositoryDelete(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

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

func TestGoCacheRepositoryExistsReturnsFalseForMissingKey(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

	if repo.Exists(ctx, "missing") {
		t.Fatal("Exists() = true, want false")
	}
}

func TestGoCacheRepositoryClear(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

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

func TestGoCacheRepositoryTTL(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

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

func TestGoCacheRepositoryTTLNoExpiration(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

	if err := repo.Set(ctx, "ttl", "value", gocache.NoExpiration); err != nil {
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

func TestGoCacheRepositoryTouch(t *testing.T) {
	ctx := context.Background()
	repo := newTestGoCacheRepository(t)

	if err := repo.Set(ctx, "touch", "value", 30*time.Millisecond); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := repo.Touch(ctx, "touch", 200*time.Millisecond); err != nil {
		t.Fatalf("Touch() error = %v", err)
	}

	time.Sleep(60 * time.Millisecond)
	if !repo.Exists(ctx, "touch") {
		t.Fatal("Exists() after Touch() = false, want true")
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
