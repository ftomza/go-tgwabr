package main

import (
	"fmt"
	"time"

	"github.com/bluele/gcache"
)

func main() {
	var evictCounter, loaderCounter, purgeCounter int
	gc := gcache.New(20).
		LRU().
		LoaderExpireFunc(func(key interface{}) (interface{}, *time.Duration, error) {
			loaderCounter++
			expire := 1 * time.Second
			return "ok", &expire, nil
		}).
		EvictedFunc(func(key, value interface{}) {
			evictCounter++
			fmt.Println("evicted key:", key)
		}).
		PurgeVisitorFunc(func(key, value interface{}) {
			purgeCounter++
			fmt.Println("purged key:", key)
		}).
		Build()
	value, err := gc.Get("key")
	if err != nil {
		panic(err)
	}
	fmt.Println("Get:", value)
	time.Sleep(1 * time.Second)
	value, err = gc.Get("key")
	if err != nil {
		panic(err)
	}
	fmt.Println("Get:", value)
	gc.Purge()
	if loaderCounter != evictCounter+purgeCounter {
		panic("bad")
	}
}
