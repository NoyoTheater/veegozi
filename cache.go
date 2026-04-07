package veegozi

import (
	"context"
	"time"

	"github.com/karlseguin/ccache/v3"
	"github.com/rs/zerolog"
)

// Cache is a generic wrapper around ccache.Cache that provides logging and simpler TTL management.
type Cache[T any] struct {
	cache *ccache.Cache[T]
	ttl   time.Duration
}

func newCache[T any](ttl time.Duration, capacity int64) *Cache[T] {
	config := ccache.Configure[T]().MaxSize(capacity)

	return &Cache[T]{
		cache: ccache.New(config),
		ttl:   ttl,
	}
}

func (c *Cache[T]) Get(ctx context.Context, key string) (T, bool) {
	log := zerolog.Ctx(ctx).With().Str("cache_key", key).Logger()

	item := c.cache.Get(key)
	if item == nil {
		log.Debug().Msg("cache miss")
		var zero T
		return zero, false
	}
	log.Debug().Msg("cache hit")
	return item.Value(), true
}

func (c *Cache[T]) Set(ctx context.Context, key string, value T) {
	log := zerolog.Ctx(ctx).With().Str("cache_key", key).Logger()
	log.Debug().Msg("setting cache entry")
	c.cache.Set(key, value, c.ttl)
}

func (c *Cache[T]) Delete(ctx context.Context, key string) {
	log := zerolog.Ctx(ctx).With().Str("cache_key", key).Logger()
	log.Debug().Msg("deleting cache entry")
	c.cache.Delete(key)
}

func (c *Cache[T]) Clear(ctx context.Context) {
	log := zerolog.Ctx(ctx).With().Logger()
	log.Debug().Msg("clearing cache")
	c.cache.Clear()
}
