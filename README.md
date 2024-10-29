# super-simple-cache
A simple cache for monoliths

## Usage

```go
type user struct {
	ID               int
	LastLoginAttempt time.Time
}

//	Use the "simplest" of caches
cache1 := simple.NewCache[int, user]()

//	Or the channel-based cache
cache2 := channel.NewCache[int, user]()

//	Store values
simple.Set(cache1, user1.ID, user1, time.Minute*1)
channel.Set(cache2, user1.ID, user1, time.Minute*1)

//	Retrieve values (these are typed!)
val1, err := simple.Get(cache1, user1.ID, nil)
val2, err := channel.Get(cache2, user1.ID, nil)

```

## Resolving a cache miss
```go
val, err := simple.Get(cache, user1.ID, func(key int) simple.CacheMissedResult[user] {
	//	A great place to call an API or access the DB!
	userFromApi, err := someApi.Call(...)
	if err != nil {
		return simple.CacheMissedResult[user]{
			Err:   err,
		}
	}

	//	Return the user we found from our API and expire it in one minute
	return simple.CacheMissedResult[user]{
		Value: userFromApi,
		TTL:   time.Minute * 1,
		Err:   nil,
	}
})
```
