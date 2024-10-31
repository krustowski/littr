package db

import (
	"sync"
)

var (
	FlowCache         Cacher
	PollCache         Cacher
	RequestCache      Cacher
	SubscriptionCache Cacher
	TokenCache        Cacher
	UserCache         Cacher
)

type Database struct {
	// Protect the stack as a whole with a proper mutex.
	mu sync.RWMutex

	// Members of the stack mush implement the Cacher interface.
	Members []Cacher
}

var database *Database

func init() {
	// Initialize all the legacy in-memory databases (caches).
	FlowCache = NewDefaultCache("FlowCache")
	PollCache = NewDefaultCache("PollCache")
	RequestCache = NewDefaultCache("RequestCache")
	SubscriptionCache = NewDefaultCache("SubscriptionCache")
	TokenCache = NewDefaultCache("TokenCache")
	UserCache = NewDefaultCache("UserCache")

	database = &Database{
		Members: []Cacher{},
	}

	// Lock the initial state.
	database.mu.Lock()

	return
}

func Lock() {
	database.mu.Lock()
}

func Unlock() {
	database.mu.Unlock()
}
