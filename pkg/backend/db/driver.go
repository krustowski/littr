package db

import (
	"go.vxn.dev/swis/v5/pkg/core"
)

var (
	FlowCache         *core.Cache
	PollCache         *core.Cache
	RequestCache      *core.Cache
	SubscriptionCache *core.Cache
	TokenCache        *core.Cache
	UserCache         *core.Cache
)

type state struct {
	unlocked bool
}

var dbState state

// Unlock function ensures that the database driver is set to the readwrite mode.
func Unlock() {
	dbState = state{unlocked: true}
}

// Lock function ensures that the database driver is set to the readonly mode.
func Lock() {
	dbState = state{unlocked: false}
}

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

// GetOne
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
	if !dbState.unlocked {
		return false
	}

	return cache.Set(key, model)
}

// DeleteOne
func DeleteOne(cache *core.Cache, key string) bool {
	if !dbState.unlocked {
		return false
	}

	return cache.Delete(key)
}
