package channel

import (
	"fmt"
	"time"
)

type Cache[K comparable, V any] struct {
	requests chan operation[K, V]
}

func NewCache[K comparable, V any]() *Cache[K, V] {

	c := Cache[K, V]{
		requests: make(chan operation[K, V]),
	}

	go listen(c.requests)

	return &c
}

func Set[K comparable, V any](cache *Cache[K, V], key K, value V, ttl time.Duration) {
	cache.requests <- operation[K, V]{
		typ:   set,
		key:   key,
		ttl:   ttl,
		value: value,
	}
}

func Get[K comparable, V any](cache *Cache[K, V], key K, resolver func(key K) CacheMissedResult[V]) (*V, error) {
	resultChan := make(chan result[V])

	cache.requests <- operation[K, V]{
		typ:        get,
		key:        key,
		resultChan: resultChan,
		resolver:   resolver,
	}

	res := <-resultChan
	close(resultChan)

	if res.err != nil {
		return nil, res.err
	}

	return res.value, nil
}

func listen[K comparable, V any](requests chan operation[K, V]) {
	store := map[K]expirableValue[V]{}

	for {
		op := <-requests // this is a block call
		switch op.typ {
		case get:
			if op.resolverError != nil {
				op.resultChan <- result[V]{
					err: op.resolverError,
				}
				continue
			}

			val, ok := store[op.key]
			if ok {
				if time.Now().Before(val.expiration) {
					op.resultChan <- result[V]{
						value: &val.value,
					}
					continue
				}

				if op.resolver == nil {
					op.resultChan <- result[V]{
						err: fmt.Errorf("key (%v) has expired", op.key),
					}
					continue
				}

			}

			if op.resolver == nil {
				op.resultChan <- result[V]{
					err: fmt.Errorf("key (%v) not found", op.key),
				}
				continue
			}

			go func(innerOp operation[K, V]) {
				resolved := innerOp.resolver(innerOp.key)
				if resolved.Err != nil {
					requests <- operation[K, V]{
						typ:           get,
						resolverError: resolved.Err,
						resultChan:    innerOp.resultChan,
					}
					return
				}
				requests <- operation[K, V]{
					typ:        set,
					key:        innerOp.key,
					value:      resolved.Value,
					ttl:        resolved.TTL,
					resolver:   innerOp.resolver,
					resultChan: innerOp.resultChan,
					relayGet:   true,
				}
			}(op)

		case set:
			store[op.key] = expirableValue[V]{
				value:      op.value,
				expiration: time.Now().Add(op.ttl),
			}

			if op.relayGet {
				go func() {
					requests <- operation[K, V]{
						typ:        get,
						key:        op.key,
						resultChan: op.resultChan,
						resolver:   op.resolver,
					}
				}()
			}
		}
	}
}

type operationType uint

const (
	set operationType = 0
	get operationType = 1
)

type operation[K comparable, V any] struct {
	typ           operationType
	key           K
	value         V
	ttl           time.Duration
	resolver      func(key K) CacheMissedResult[V]
	resolverError error
	relayGet      bool
	resultChan    chan result[V]
}

type result[V any] struct {
	value *V
	err   error
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
