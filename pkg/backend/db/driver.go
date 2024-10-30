package db

import (
	"reflect"
	"sync"

	"go.vxn.dev/littr/pkg/backend/metrics"
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

// state is a helper structure to represent the current state of the databases.
type state struct {
	unlocked bool
	loaded   bool
}

var dbState *state

// Unlock function ensures that the database driver is set to the readwrite (RW) mode.
func Unlock() {
	if dbState == nil {
		dbState = &state{unlocked: true}
	} else {
		dbState.unlocked = true
	}
}

// Lock function ensures that the database driver is set to the readonly (RO) mode.
func Lock() {
	if dbState == nil {
		dbState = &state{unlocked: false}
	} else {
		dbState.unlocked = false
	}
}

// MarkLoaded function unlock the database(s) state for the assiciated metrics updating. This means that the database has been loaded on start.
func MarkLoaded() {
	if dbState == nil {
		dbState = &state{loaded: true}
	} else {
		dbState.loaded = true
	}
}

// Mutex should ensure the combined read+write operations (SetOne here specially) are thread safe.
// https://stackoverflow.com/a/66774210
var mu sync.Mutex

// GetAll function does a fetch-all operation on the given cache instance. After the data retrieval, all items are to be asserted their corresponding types and to be loaded into a map[string]T map, where T is a generic type. This map and number of processed items are returned.
func GetAll[T any](cache *core.Cache, model T) (map[string]T, int) {
	// An initialization check.
	if cache == nil {
		return map[string]T{}, 0
	}

	// Fetch all items as the map[string]interface{} map.
	itemsInterface, count := cache.GetAll()

	// Prepare the data KV map.
	items := make(map[string]T)

	var control int

	// Loop over all keyed interfaces and assert type T to every one of them, compose an output map. Make a control counting.
	for key, rawItem := range itemsInterface {
		control++

		// Assert the specific type.
		item, ok := rawItem.(T)
		if !ok {
			// Continue rather than exiting, compare the count values afterwards.
			//return items, count
			continue
		}
		items[key] = item
	}

	// Write the count to the associated metric.
	metrics.UpdateCountMetric(cache, count, true)

	// Return the lower count, because it reflects the actual valid items count.
	if control < count {
		return items, control
	}
	return items, count
}

// GetOne fetches just one very item from the given cache instance. As long as the function is generic, the type is asserted automatically, so the type passing is required. Returns the requested item and the its retrieval result as boolean.
func GetOne[T any](cache *core.Cache, key string, model T) (T, bool) {
	// An initialization check.
	if cache == nil {
		return model, false
	}

	// Fetch the requested item from the cache.
	itemInterface, found := cache.Get(key)
	if !found {
		return model, false
	}

	// Assert the given type T.
	item, ok := itemInterface.(T)
	if !ok {
		return model, false
	}

	// Return the item of its asserted type.
	return item, true
}

// SetOne writes the input value of some type to the corresponding cache storing the very same item type (TODO ensure this).
// Fails if the database state is locked or uninitialized. Please note that this very function has to have another sync mechanism implemented,
// as the combined read+write operation is not considered a thread safe.
func SetOne[T any](cache *core.Cache, key string, model T) bool {
	// An initialization check.
	if !dbState.unlocked || cache == nil {
		return false
	}

	var doIncrementMetric bool = true

	// Lock the mutex.
	mu.Lock()
	defer mu.Unlock()

	// Check for the possible item's existence in such cache instance. The item will be rewritten anyway (unless),
	// but this is to make sure we are not incrementing the statistics while the count remains the same.
	control, found := GetOne[T](cache, key, model)
	if found {
		doIncrementMetric = !found

		// Check if the control item is deeply equal with the requested item to save. If so, do not save new item and unlock the mutex.
		if reflect.DeepEqual(control, model) {
			return true
		}
	}

	// Load the item to the corresponding key into the cache.
	saved := cache.Set(key, model)

	// Increment the metric count only if the item is a new one, has been saved properly, and the database state is fully loaded.
	if saved && dbState.loaded && doIncrementMetric {
		metrics.UpdateCountMetric(cache, 1, false)
	}

	return saved
}

// DeleteOne deletes an item from such cache via the requested key value. Fails if the database state is locked.
func DeleteOne(cache *core.Cache, key string) bool {
	// The database state check.
	if !dbState.unlocked {
		return false
	}

	var deleted bool

	// Delete the item using its key.
	if deleted = cache.Delete(key); deleted {
		// Update the metric count when deleted properly.
		metrics.UpdateCountMetric(cache, -1, false)
	}

	return deleted
}
