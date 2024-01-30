package backend

import (
	"go.savla.dev/swis/v5/pkg/core"
)

var (
	FlowCache         *core.Cache
	PollCache         *core.Cache
	SubscriptionCache *core.Cache
	TimestampCache    *core.Cache
	UserCache         *core.Cache
)

// getAll
func getAll[T any](cache *core.Cache, model T) (map[string]T, int) {
	itemsInterface, count := cache.GetAll()

	items := make(map[string]T)

	for key, rawItem := range itemsInterface {
		item, ok := rawItem.(T)
		if !ok {
			return items, count
		}
		items[key] = item
	}

	return items, count
}

// getMany
// seek and skip implementation TBD --- pagination idea
// mark being the ID/key, stop/start the export there in the cache map
// export the 'count' number of records in map with IDs/keys
// the 'count' number passthrough to the 2nd return value
func getMany[T any](cache *core.Cache, model T, pagination int, page int) (map[string]T, int) {
	items := make(map[string]T)

	// keys (and models) array acts like a helper array for the further proper item cut and export
	keys := []string{}
	//models := []T{}

	// let us fetch the data
	itemsInterface, _ := cache.GetAll()

	// modify the range contents of the getAll function
	for key, rawItem := range itemsInterface {
		item, ok := rawItem.(T)
		if !ok {
			// terminate the cycle on type assertion error
			// maybe the corrupted data could be skipped to keep the cycle running
			//return items, 0
			//break
			continue
		}

		items[key] = item
		keys = append(keys, key)
		//models = append(models, item)
	}

	// reverse the order if requested
	// https://stackoverflow.com/a/19239850
	reverse(keys)

	// return one page of items
	var toExport map[string]T = make(map[string]T)
	//for i := len(keys) - 1 - (pagination * (page - 1)); i > len(keys)-(pagination*page) && i > 0; i-- {
	for i := pagination * (page - 1); i < pagination*page; i++ {
		toExport[keys[i]] = items[keys[i]]
	}

	return toExport, pagination
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

func deleteOne(cache *core.Cache, key string) bool {
	return cache.Delete(key)
}
