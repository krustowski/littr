package backend

import (
	"go.savla.dev/swis/v5/pkg/core"
)

var (
	FlowCache    *core.Cache
	PollCache    *core.Cache
	SessionCache *core.Cache
	UserCache    *core.Cache
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

// getMany
// seek and skip implementation TBD --- pagination idea
// mark being the ID/key, stop/start the export there in the cache map
// export the 'count' number of records in map with IDs/keys
// the 'count' number passthrough to the 2nd return value
// 'reverse' bool can be used to
func getMany[T any](cache *core.Cache, model T, mark string, countMany int, reverse bool) (map[string]T, int) {
	items := make(map[string]T)

	// keys (and models) array acts like a helper array for the further proper item cut and export
	keys := []string{}
	models := []T{}

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

		keys = append(keys, key)
		models = append(models, item)

		// check if we hit the marked position in the map
		if key == mark {
			break
		}
	}

	// reverse the order if requested
	// https://stackoverflow.com/a/19239850
	if reverse {
		for i, j := 0, len(models)-1; i < j; i, j = i+1, j-1 {
			keys[i], keys[j] = keys[j], keys[i]
			models[i], models[j] = models[j], models[i]
		}
	}

	// compose the range for scissors
	start := 0
	stop := countMany - 1

	// cut off the requested number of items
	if stop < len(keys) {
		keysCut := keys[start:stop]
		modelsCut := models[start:stop]

		keys = keysCut
		models = modelsCut
	}

	// reassemble a map to return
	for idx, key := range keys {
		items[key] = models[idx]
	}

	return items, len(keys)
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
