package cache

import (
	"fmt"
	"time"
	"weather-lambda/internal/log"

	"github.com/patrickmn/go-cache"
)

var c = cache.New(5*time.Minute, 10*time.Minute)

func SetCache(key string, value interface{}) {
	log.Info(fmt.Sprintf("Setting cache for key: %s", key))
	c.Set(key, value, cache.DefaultExpiration)
}

func GetCache(key string) (interface{}, bool) {
	data, found := c.Get(key)
	if found {
		log.Info(fmt.Sprintf("Cache hit for key: %s", key))
	} else {
		log.Info(fmt.Sprintf("Cache miss for key: %s", key))
	}
	return data, found
}
