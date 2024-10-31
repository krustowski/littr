package main

import (
	"fmt"
	rn "math/rand/v2"
	"sync"
	"testing"
)

// Dummy is a dummy struct acting like the generic cache's payload (a value associated with a key).
type Dummy struct {
	Name   string
	Number int
	//Float  float64
	Keys []string
}

//
//  Mutex + pointer map???
//

type mutexMap struct {
	mu sync.Mutex
	m  map[string]Dummy
}

func NewMutexMap() *mutexMap {
	return &mutexMap{m: make(map[string]Dummy)}
}

func (m *mutexMap) Load(key string) Dummy {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.m[key]
}

func (m *mutexMap) Store(key string, val Dummy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = val
	return
}

func TestLoadStoreMutex(t *testing.T) {}

func BenchmarkLoadStoreMutex(b *testing.B) {
	var wg sync.WaitGroup
	var m *mutexMap = NewMutexMap()

	var lambda = make([]func(K string, w *sync.WaitGroup), 2)
	lambda[0] = func(K string, w *sync.WaitGroup) {
		defer w.Done()

		_ = m.Load(K)
	}
	lambda[1] = func(K string, w *sync.WaitGroup) {
		defer w.Done()

		//V := Dummy{Name: K, Number: rand.IntN(100), Float: rand.FloatN(100)}
		V := Dummy{Name: K, Number: rn.IntN(100)}

		m.Store(K, V)
	}

	var keys = make([]string, b.N)

	fmt.Printf("Starting the benchmark (N=%d)...\n\n", b.N)

	wg.Add(b.N)

	// Start the timer and run the banchmark's core.
	//start := time.Now()
	for i := 0; i < b.N; i++ {
		// Randomly assign the Load/Store lambda, geenrate a key
		grp := rn.IntN(1)
		keys[i] = RandStringBytesMaskImprSrcUnsafe(5)

		// Run a single of many goroutines.
		go lambda[grp](keys[i], &wg)
	}

	wg.Wait()
	//stop := time.Now().Sub(start)

	//fmt.Printf("Stop: %.3f ops/ns \t %.3f ns/op \n", float64(int64(b.N)/int64(stop)), float64(int64(stop)/int64(b.N)))
}

//
//  RWMutex
//

type rwmutexMap struct {
	mu sync.RWMutex
	m  map[string]Dummy
}

func NewRWMutexMap() *rwmutexMap {
	return &rwmutexMap{m: make(map[string]Dummy)}
}

func (m *rwmutexMap) Load(key string) Dummy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.m[key]
}

func (m *rwmutexMap) Store(key string, val Dummy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = val
	return
}

func TestLoadStoreRWMutex(t *testing.T) {}

func BenchmarkLoadStoreRWMutex(b *testing.B) {
	var wg sync.WaitGroup
	var m *rwmutexMap = NewRWMutexMap()

	var lambda = make([]func(K string, w *sync.WaitGroup), 2)
	lambda[0] = func(K string, w *sync.WaitGroup) {
		defer w.Done()

		_ = m.Load(K)
	}
	lambda[1] = func(K string, w *sync.WaitGroup) {
		defer w.Done()

		//V := Dummy{Name: K, Number: rand.IntN(100), Float: rand.FloatN(100)}
		V := Dummy{Name: K, Number: rn.IntN(100)}

		m.Store(K, V)
	}

	var keys = make([]string, b.N)

	//fmt.Printf("Starting the benchmark (N=%d)...\n\n", b.N)

	wg.Add(b.N)

	// Start the timer and run the banchmark's core.
	//start := time.Now()
	for i := 0; i < b.N; i++ {
		// Randomly assign the Load/Store lambda, geenrate a key
		grp := rn.IntN(1)
		keys[i] = RandStringBytesMaskImprSrcUnsafe(5)

		// Run a single of many goroutines.
		go lambda[grp](keys[i], &wg)
	}

	wg.Wait()
	//stop := time.Now().Sub(start)

	//fmt.Printf("Stop: %.3f ops/ns \t %.3f ns/op \n", float64(int64(b.N)/int64(stop)), float64(int64(stop)/int64(b.N)))
}

//
//  syncMap
//

type syncMap struct {
	m sync.Map
}

func NewSyncMap() *syncMap {
	return &syncMap{}
}

func (m *syncMap) Load(key string) {
	m.Load(key)
}

func (m *syncMap) Store(key string, val interface{}) {
	m.Store(key, val)
}

func BenchmarkLoadStoreSyncMap(b *testing.B) {
	var wg sync.WaitGroup
	var m *syncMap = NewSyncMap()

	var lambda = make([]func(K string, w *sync.WaitGroup), 2)
	lambda[0] = func(K string, w *sync.WaitGroup) {
		defer w.Done()

		m.Load(K)
	}
	lambda[1] = func(K string, w *sync.WaitGroup) {
		defer w.Done()

		//V := Dummy{Name: K, Number: rand.IntN(100), Float: rand.FloatN(100)}
		V := Dummy{Name: K, Number: rn.IntN(100)}

		m.Store(K, V)
	}

	var keys = make([]string, b.N)

	//fmt.Printf("Starting the benchmark (N=%d)...\n\n", b.N)

	wg.Add(b.N)

	// Start the timer and run the banchmark's core.
	//start := time.Now()
	for i := 0; i < b.N; i++ {
		// Randomly assign the Load/Store lambda, geenrate a key
		grp := rn.IntN(1)
		keys[i] = RandStringBytesMaskImprSrcUnsafe(5)

		// Run a single of many goroutines.
		go lambda[grp](keys[i], &wg)
	}

	wg.Wait()
	//stop := time.Now().Sub(start)

	//fmt.Printf("Stop: %.3f ops/ns \t %.3f ns/op \n", float64(int64(b.N)/int64(stop)), float64(int64(stop)/int64(b.N)))
}
