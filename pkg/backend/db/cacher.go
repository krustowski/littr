package db

import (
	"sync"
)

type Cacher interface {
	// Load takes a key string and fetches its value from the implementing cache store.
	// Returns true if the item has been found, otherwise writes false.
	Load(key string) (interface{}, bool)

	// Store takes a key-value pair and stores the value with the associated key into the implementing cache store.
	// Simple overwrite.
	Store(key string, value interface{})

	// Delete takes a key string and deletes its record in the implementing cache store.
	Delete(key string)

	// Range loop over all the implementing cache keys and retrieves all the associated values as interface{}.
	// Returns the pointer to a GenericMap with all keys and their associated values, and the item count.
	Range() (*GenericMap, int64)
}

type GenericMap map[string]interface{}

//
//  SimpleCache
//  Simple implementation of the Cacher interface.
//

type SimpleCache struct {
	// Name
	Name string

	// Generic string-keyed map protected by Mutex.
	mu sync.Mutex
	mp GenericMap
}

func NewSimpleCache(name string) *SimpleCache {
	if name == "" {
		return nil
	}

	return &SimpleCache{
		Name: name,
		mp:   make(GenericMap),
	}
}

func (c *SimpleCache) Load(key string) (interface{}, bool) {
	c.mu.Lock()
	rawV := c.mp[key]
	c.mu.Unlock()

	if rawV == nil {
		return nil, false
	}

	return rawV, true
}

func (c *SimpleCache) Store(key string, rawV interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mp[key] = rawV
	return
}

func (c *SimpleCache) Range() (*GenericMap, int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var count int64
	var genericMap = GenericMap{}

	// Very expensive operation.
	for key, rawV := range c.mp {
		genericMap[key] = rawV
	}

	return &genericMap, count
}

//
//  SignalCache
//  Cacher interface implementation with the sync.Cond usage to notify the readers.
//  In this implementation the main purpose is to update the cache and notify waiting readers.
//

type SignalCache struct {
	// Name of the cache.
	Name string

	// Generic string-keyed map protected by a condition guarded by Cond (with RWMutex).
	c       *sync.Cond
	updated bool
	mp      GenericMap
}

func NewSignalCache(name string) *SignalCache {
	if name == "" {
		return nil
	}

	var mu sync.RWMutex

	return &SignalCache{
		// Signal cache's very name.
		Name: name,

		// Internal sync logic.
		c:       sync.NewCond(&mu),
		updated: false,
		mp:      make(GenericMap),
	}
}

func (c *SignalCache) Load(key string) (interface{}, bool) {
	c.c.L.Lock()
	for !c.updated {
		c.c.Wait()
	}

	rawV := c.mp[key]
	c.c.L.Unlock()

	if rawV == nil {
		return nil, false
	}

	return rawV, true
}

func (c *SignalCache) makeWaitForUpdate() {
	c.c.L.Lock()
	c.updated = false
	c.c.L.Unlock()

}

func (c *SignalCache) Store(key string, rawV interface{}) {
	defer c.c.Broadcast()
	c.c.L.Lock()
	c.mp[key] = rawV
	c.updated = true
	c.c.L.Unlock()
}

func (c *SignalCache) Delete(key string) {}

func (c *SignalCache) Range() (*GenericMap, int64) {
	c.c.L.Lock()
	defer c.c.L.Unlock()

	// Wait for the update.
	for !c.updated {
		c.c.Wait()
	}

	var count int64
	var genericMap = GenericMap{}

	// Very expensive operation.
	for key, rawV := range c.mp {
		genericMap[key] = rawV
	}

	return &genericMap, count
}

//
//  New main db pkg's export functions.
//

/*func GetOne[T any](cache Cacher, key string) (*T, bool) {
	rawV := cache.Load(key)
	val, ok := rawV.(*T)
	if !ok {
		return nil, false
	}

	return val, true
}

func SetOne(cache Cacher, key string, val interface{}) bool
func DeleteOne(cache Cacher, key string) bool*/
