package db

import (
	"go.savla.dev/swis/v5/pkg/core"
)

var (
	FlowCache         *core.Cache
	PollCache         *core.Cache
	RequestCache      *core.Cache
	SubscriptionCache *core.Cache
	TokenCache        *core.Cache
	UserCache         *core.Cache
)

// GetAll
func GetAll[T any](cache *core.Cache, model T) (map[string]T, int) {
	itemsInterface, count := cache.GetAll()

	items := make(map[string]T)

	// loop over all key'd interfaces and assert type T to every of them, compose a map
	for key, rawItem := range itemsInterface {
		item, ok := rawItem.(T)
		if !ok {
			return items, count
		}
		items[key] = item
	}

	return items, count
}

// etOne
func GetOne[T any](cache *core.Cache, key string, model T) (T, bool) {
	itemInterface, found := cache.Get(key)
	if !found {
		return model, false
	}

	item, ok := itemInterface.(T)
	if !ok {
		return model, false
	}

	return item, true
}

// SetOne
func SetOne[T any](cache *core.Cache, key string, model T) bool {
	return cache.Set(key, model)
}

// DeleteOne
func DeleteOne(cache *core.Cache, key string) bool {
	return cache.Delete(key)
}
