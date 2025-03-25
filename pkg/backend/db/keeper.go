package db

import "sync"

type DatabaseKeeper interface {
	ReadLock()
	ReadUnlock()
	Lock()
	Unlock()
	ReleaseLock()

	RunMigrations() (report string, err error)

	DumpAll() (report string, err error)
	LoadAll() (report string, err error)

	Database() map[string]Cacher
}

type defaultDatabaseKeeper struct {
	// Protect the stack as a whole with a proper mutex.
	mu *sync.RWMutex

	readonly bool

	caches []Cacher
}

func NewDatabase() DatabaseKeeper {
	var (
		caches []Cacher
		mu     sync.RWMutex
	)

	names := []string{
		"FlowCache",
		"PollCache",
		"RequestCache",
		"TokenCache",
		"UserCache",
	}

	for _, name := range names {
		caches = append(caches, NewDefaultCache(name))
	}

	return &defaultDatabaseKeeper{
		mu:     &mu,
		caches: caches,
	}
}

func (d *defaultDatabaseKeeper) ReadLock() {
	d.mu.RLock()
}

func (d *defaultDatabaseKeeper) ReadUnlock() {
	d.mu.RUnlock()
}

func (d *defaultDatabaseKeeper) Lock() {
	d.mu.Lock()
}

func (d *defaultDatabaseKeeper) Unlock() {
	d.mu.Unlock()
}

func (d *defaultDatabaseKeeper) ReleaseLock() {
	d.mu.Unlock()
	d.readonly = true
}

func (d *defaultDatabaseKeeper) Database() map[string]Cacher {
	m := make(map[string]Cacher)

	for _, cache := range d.caches {
		if name := cache.GetName(); name == "" {
			continue
		} else {
			m[name] = cache
		}
	}

	return m
}
