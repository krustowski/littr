package backend

import (
	"go.savla.dev/swis/v5/pkg/core"
)

var (
	FlowCache *core.Cache
	PollCache *core.Cache
	//SessionCache *core.Cache
	UserCache *core.Cache
)

func getAll[T any](cache *core.Cache, model T) (map[string]T, int) {
	itemsInterface, count := cache.GetAll()

	items := make(map[string]T)

	for key, rawItem := range itemsInterface {
		item, ok := rawItem.(T)
		if !ok {
			return items, 0
		}
		items[key] = item
	}

	return items, count
}

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

func setOne[T any](cache *core.Cache, key string, model T) bool {
	return cache.Set(key, model)
}

func DeleteOne[T any](cache *core.Cache, key string) bool {
	return cache.Delete(key)
}
