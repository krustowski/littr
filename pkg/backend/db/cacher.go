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
	Store(key string, value interface{}) bool

	// Delete takes a key string and deletes its record in the implementing cache store.
	Delete(key string) bool

	// Range loop over all the implementing cache keys and retrieves all the associated values as interface{}.
	// Returns the pointer to a GenericMap with all keys and their associated values, and the item count.
	Range() (*GenericMap, int64)

	// GetName returns the implemented cache's name.
	GetName() string
}

type GenericMap map[string]interface{}

//
//  DefaultCache
//  The legacy implementation of the Cacher interface. The original implementation (not implementing the Cacher interface) comes from swapi/core.Cache.
//

type DefaultCache struct {
	// Name of the cache.
	Name string

	syncMap sync.Map
}

func NewDefaultCache(name string) *DefaultCache {
	if name == "" {
		return nil
	}

	return &DefaultCache{
		Name: name,
	}
}

func (c *DefaultCache) Load(key string) (interface{}, bool) {
	return c.syncMap.Load(key)
}

func (c *DefaultCache) Store(key string, rawV interface{}) bool {
	c.syncMap.Store(key, rawV)
	return true
}

func (c *DefaultCache) Delete(key string) bool {
	c.syncMap.Delete(key)
	return true
}

func (c *DefaultCache) Range() (*GenericMap, int64) {
	var genericMap = GenericMap{}
	var counter int64

	c.syncMap.Range(func(rawK, rawV interface{}) bool {
		key, ok := rawK.(string)
		if !ok {
			return false
		}

		genericMap[key] = rawV
		counter++
		return true
	})

	return &genericMap, counter
}

func (c *DefaultCache) GetName() string {
	return c.Name
}

//
//  SimpleCache
//  Simple implementation of the Cacher interface. SimpleCache's map is protected by RWMutex.
//

type SimpleCache struct {
	// Name of the cache.
	Name string

	// Generic string-keyed map protected by RWMutex.
	mu sync.RWMutex
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
	c.mu.RLock()
	rawV := c.mp[key]
	c.mu.RUnlock()

	if rawV == nil {
		return nil, false
	}

	return rawV, true
}

func (c *SimpleCache) Store(key string, rawV interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mp[key] = rawV

	return true
}

func (c *SimpleCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.mp, key)

	return true
}

func (c *SimpleCache) Range() (*GenericMap, int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var counter int64
	var genericMap = GenericMap{}

	// Very expensive operation.
	for key, rawV := range c.mp {
		genericMap[key] = rawV
		counter++
	}

	return &genericMap, counter
}

func (c *SimpleCache) GetName() string {
	return c.Name
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

// makeWaitForUpdate is a metafunction to lock the condition.
func (c *SignalCache) makeWaitForUpdate() {
	c.c.L.Lock()
	c.updated = false
	c.c.L.Unlock()
}

func (c *SignalCache) Store(key string, rawV interface{}) bool {
	// Make readers wait for an update.
	c.makeWaitForUpdate()

	// Lock the write access and prepare the Unlock and Broadcast defers.
	c.c.L.Lock()
	defer c.c.L.Unlock()
	defer c.c.Broadcast()

	// Store/rewrite the value.
	c.mp[key] = rawV
	c.updated = true

	return true
}

func (c *SignalCache) Delete(key string) bool {
	// Make readers wait for an update.
	c.makeWaitForUpdate()

	// Lock the write access and prepare the Unlock and Broadcast defers.
	c.c.L.Lock()
	defer c.c.L.Unlock()
	defer c.c.Broadcast()

	// Delete the value associated with such key.
	delete(c.mp, key)

	return true
}

func (c *SignalCache) Range() (*GenericMap, int64) {
	c.c.L.Lock()
	defer c.c.L.Unlock()

	// Wait for the update.
	for !c.updated {
		c.c.Wait()
	}

	var counter int64
	var genericMap = GenericMap{}

	// Very expensive operation.
	for key, rawV := range c.mp {
		genericMap[key] = rawV
		counter++
	}

	return &genericMap, counter
}

func (c *SignalCache) GetName() string {
	return c.Name
}
