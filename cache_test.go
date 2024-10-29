package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Thawe/super-simple-cache/channel"
	"github.com/Thawe/super-simple-cache/simple"
	"golang.org/x/sync/errgroup"
)

func BenchmarkCache(b *testing.B) {

	type user struct {
		ID               int
		LastLoginAttempt time.Time
	}

	b.Run("simple", func(b *testing.B) {
		cache := simple.NewCache[int, user]()
		g, _ := errgroup.WithContext(context.Background())

		process := func(userID int) error {
			now := time.Now()
			user1 := user{
				ID:               userID,
				LastLoginAttempt: now,
			}

			simple.Set(cache, user1.ID, user1, time.Minute*1)

			val, err := simple.Get(cache, user1.ID, nil)
			if err != nil {
				return err
			} else {
				if !reflect.DeepEqual(user1, *val) {
					return fmt.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
				}
			}

			return nil
		}

		b.Logf("count: %d", b.N)
		j := 0
		for i := 0; i < b.N; i++ {
			for j%10000 != 0 {
				userID := j
				g.Go(func() error {
					return process(userID)
				})
				j++
			}
			j++
		}

		err := g.Wait()
		if err != nil {
			b.Error(err)
		}
	})

	b.Run("channel cache", func(b *testing.B) {
		cache := channel.NewCache[int, user]()
		g, _ := errgroup.WithContext(context.Background())

		process := func(userID int) error {
			now := time.Now()
			user1 := user{
				ID:               userID,
				LastLoginAttempt: now,
			}

			channel.Set(cache, user1.ID, user1, time.Minute*1)

			val, err := channel.Get(cache, user1.ID, nil)
			if err != nil {
				return err
			} else {
				if !reflect.DeepEqual(user1, *val) {
					return fmt.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
				}
			}

			return nil
		}

		b.Logf("count: %d", b.N)
		j := 0
		for i := 0; i < b.N; i++ {
			for j%10000 != 0 {
				userID := j
				g.Go(func() error {
					return process(userID)
				})
				j++
			}
			j++
		}

		err := g.Wait()
		if err != nil {
			b.Error(err)
		}
	})
}

func TestCacheDoesExpire(t *testing.T) {

	type user struct {
		ID               int
		LastLoginAttempt time.Time
	}

	now := time.Now()
	user1 := user{
		ID:               1,
		LastLoginAttempt: now,
	}

	t.Run("simple", func(t *testing.T) {
		cache := simple.NewCache[int, user]()

		simple.Set(cache, user1.ID, user1, time.Second*1)

		val, err := simple.Get(cache, user1.ID, nil)
		if err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(user1, *val) {
				t.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
			}
		}

		time.Sleep(time.Second * 1)

		_, err = simple.Get(cache, user1.ID, nil)
		if err == nil {
			t.Errorf("expected key to expire")
		}
	})

	t.Run("channel cache", func(t *testing.T) {
		cache := channel.NewCache[int, user]()

		channel.Set(cache, user1.ID, user1, time.Second*1)

		val, err := channel.Get(cache, user1.ID, nil)
		if err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(user1, *val) {
				t.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
			}
		}

		time.Sleep(time.Second * 1)

		_, err = channel.Get(cache, user1.ID, nil)
		if err == nil {
			t.Errorf("expected key to expire")
		}
	})
}

func TestCacheResolver(t *testing.T) {

	type user struct {
		ID               int
		LastLoginAttempt time.Time
	}

	now := time.Now()
	user1 := user{
		ID:               1,
		LastLoginAttempt: now,
	}

	t.Run("simple", func(t *testing.T) {
		cache := simple.NewCache[int, user]()

		val, err := simple.Get(cache, user1.ID, func(key int) simple.CacheMissedResult[user] {
			time.Sleep(time.Millisecond * 500)
			return simple.CacheMissedResult[user]{
				Value: user1,
				TTL:   time.Minute * 1,
				Err:   nil,
			}
		})

		if err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(user1, *val) {
				t.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
			}
		}

		again, err := simple.Get(cache, user1.ID, nil)
		if err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(user1, *again) {
				t.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
			}
		}

	})

	t.Run("channel cache", func(t *testing.T) {
		cache := channel.NewCache[int, user]()

		val, err := channel.Get(cache, user1.ID, func(key int) channel.CacheMissedResult[user] {
			time.Sleep(time.Millisecond * 500)
			return channel.CacheMissedResult[user]{
				Value: user1,
				TTL:   time.Minute * 1,
				Err:   nil,
			}
		})

		if err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(user1, *val) {
				t.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
			}
		}

		again, err := channel.Get(cache, user1.ID, nil)
		if err != nil {
			t.Error(err)
		} else {
			if !reflect.DeepEqual(user1, *again) {
				t.Errorf("\nexpected (%v)\ngot      (%v)", user1, *val)
			}
		}

	})

}
