package simple

import (
	"fmt"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	store map[K]expirableValue[V]
	mu    sync.Mutex
}

func NewCache[K comparable, V any]() *Cache[K, V] {

	c := Cache[K, V]{
		store: map[K]expirableValue[V]{},
		mu:    sync.Mutex{},
	}

	return &c
}

func Set[K comparable, V any](cache *Cache[K, V], key K, value V, ttl time.Duration) {
	cache.mu.Lock()
	cache.store[key] = expirableValue[V]{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
	cache.mu.Unlock()
}

func Get[K comparable, V any](cache *Cache[K, V], key K, resolver func(key K) CacheMissedResult[V]) (*V, error) {
	cache.mu.Lock()
	val, ok := cache.store[key]
	cache.mu.Unlock()

	if ok {
		if time.Now().Before(val.expiration) {
			return &val.value, nil
		}

		if resolver == nil {
			return nil, fmt.Errorf("key (%v) has expired", key)
		}
	}

	if resolver == nil {
		return nil, fmt.Errorf("key (%v) not found", key)
	}

	resolved := resolver(key)
	if resolved.Err != nil {
		return nil, resolved.Err
	}

	Set(cache, key, resolved.Value, resolved.TTL)
	return &resolved.Value, nil
}

type expirableValue[V any] struct {
	value      V
	expiration time.Time
}

type CacheMissedResult[T any] struct {
	Value T
	TTL   time.Duration
	Err   error
}
