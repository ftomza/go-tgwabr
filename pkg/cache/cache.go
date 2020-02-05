package cache

import (
	"context"
	"fmt"
	"tgwabr/api"
	"time"

	"github.com/bluele/gcache"
)

type Cache struct {
	ctx        context.Context
	gc         gcache.Cache
	getMembers func() ([]int, error)

	api.Cache
}

type Config struct {
	GetMembers func() ([]int, error)
}

func (s *Cache) loaderExpireFunc(key interface{}) (value interface{}, duration *time.Duration, err error) {
	switch key.(type) {
	case string:
		switch key {
		case "members":
			if s.getMembers == nil {
				return nil, nil, fmt.Errorf("Handler getMembers not define ")
			}
			value, err = s.getMembers()
			exp := 120 * time.Second
			return value, &exp, err
		default:
			return nil, nil, fmt.Errorf("Key '%s' not define ", key)
		}
	default:
		return nil, nil, fmt.Errorf("Type key '%T' not define ", key)
	}
}

func New(ctx context.Context, config Config) (cache *Cache, err error) {

	cache = &Cache{ctx: ctx, getMembers: config.GetMembers}

	cache.gc = gcache.New(20).
		LRU().
		LoaderExpireFunc(cache.loaderExpireFunc).
		Build()

	return
}

func (s *Cache) ShutDown() error {
	s.gc.Purge()
	return nil
}
