package cache

import (
	"log"
	"testing"
)

func TestRepositoriesUseDefaultLogger(t *testing.T) {
	goCache := NewGoCacheRepository(0, 0, nil).(*GoCacheRepository)
	freeCache := NewFreecacheRepository(1024*1024, nil).(*FreecacheRepository)
	goRedis := NewGoRedisRepository(nil, nil).(*GoRedisRepository)

	for name, logger := range map[string]*log.Logger{
		"GoCacheRepository":   goCache.Logger,
		"FreecacheRepository": freeCache.Logger,
		"GoRedisRepository":   goRedis.Logger,
	} {
		if logger != log.Default() {
			t.Errorf("%s Logger is not log.Default()", name)
		}
	}
}
