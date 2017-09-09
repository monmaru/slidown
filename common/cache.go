package common

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

// SetCache ...
func SetCache(ctx context.Context, key string, in interface{}) {
	item := &memcache.Item{
		Key:        key,
		Object:     in,
		Expiration: 1 * time.Hour,
	}
	if err := memcache.Gob.Set(ctx, item); err != nil {
		log.Errorf(ctx, "error setting item: %v", err)
	}
}

// GetCache ...
func GetCache(ctx context.Context, key string, in interface{}) error {
	if _, err := memcache.Gob.Get(ctx, key, in); err == memcache.ErrCacheMiss {
		log.Infof(ctx, "item not in the cache")
		return err
	} else if err != nil {
		log.Errorf(ctx, "error getting item: %v", err)
		return err
	}

	log.Debugf(ctx, "cache hit: %v", key)
	return nil
}
