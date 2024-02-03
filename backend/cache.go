package backend

import (
	"go.savla.dev/swis/v5/pkg/core"
)

var (
	FlowCache         *core.Cache
	PollCache         *core.Cache
	SubscriptionCache *core.Cache
	UserCache         *core.Cache
)

// getAll
func getAll[T any](cache *core.Cache, model T) (map[string]T, int) {
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

// getOne
func getOne[T any](cache *core.Cache, key string, model T) (T, bool) {
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

// setOne
func setOne[T any](cache *core.Cache, key string, model T) bool {
	return cache.Set(key, model)
}

// deleteOne
func deleteOne(cache *core.Cache, key string) bool {
	return cache.Delete(key)
}
