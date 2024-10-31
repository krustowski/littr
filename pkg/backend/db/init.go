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

	// Helper booleans to simply track the stack's state.
	RLocked bool
	Locked  bool

	// Members of the stack mush implement the Cacher interface.
	Members []Cacher
}

var database *Database

func init() {
	// Initialize all the legacy in-memory databases (caches).
	flowCache := NewDefaultCache("FlowCache")
	pollCache := NewDefaultCache("PollCache")
	requestCache := NewDefaultCache("RequestCache")
	subscriptionCache := NewDefaultCache("SubscriptionCache")
	tokenCache := NewDefaultCache("TokenCache")
	userCache := NewDefaultCache("UserCache")

	// Update the main pkg-exported pointers.
	FlowCache = flowCache
	PollCache = pollCache
	RequestCache = requestCache
	SubscriptionCache = subscriptionCache
	TokenCache = tokenCache
	UserCache = userCache

	// Pass the cache pointers to the stack.
	database = &Database{
		Members: []Cacher{
			flowCache,
			pollCache,
			requestCache,
			subscriptionCache,
			tokenCache,
			userCache,
		},
	}

	// Explicitly state the defaults.
	database.Locked = false
	database.RLocked = false

	// RLock the initial RO state, so nobody can read before the data are imported and migrated properly.
	database.mu.RLock()

	return
}

//
//  Lockers for the database stack.
//

func Lock() {
	database.mu.Lock()
	database.Locked = true
}

func RLock() {
	database.mu.RLock()
	database.RLocked = true
}

func Unlock() {
	database.mu.Unlock()
	database.Locked = false
}

func RUnlock() {
	database.mu.RUnlock()
	database.RLocked = false
}

// Release is a special fuction evoked when the lock needs to be released before the backend shutdown not to defer its exit, but to keep the database stack in the read-only state.
func ReleaseLock() {
	database.mu.Unlock()
	database.Locked = true
}
