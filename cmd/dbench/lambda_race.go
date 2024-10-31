package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"go.vxn.dev/littr/pkg/backend/db"
)

//
//  TYPES
//

type testConfiguration struct {
	// Bank (array) of various Cacher interface implementations.
	CacheBank []db.Cacher

	// Keys are the key component of a cache = the very data are linked to various keys as so-called key-value pairs.
	Keys []string

	// A stable key length.
	KeyLen int

	// The key count, number of keys to be generated.
	KeyCount int

	// TestChannels filed is to hold the all the channels created at the test start.
	TestChannels []chan interface{}

	// TestOperations is an array of lambda functions of type testPperationFunc.
	TestOperations     []testOperationFunc
	TestOperationNames []string

	// Timeout is to race with the sync's WaitGroup, whoever would make it first from the testing go routine.
	// Timeout is used by the timeoutHandler lambda.
	Timeout time.Duration

	// The number of workers to be spinned off.
	WorkerCount int

	// An indication whether the worker count should be a random number.
	WorkerCountRandom bool

	// The <0,N) range for the random integer generator.
	WorkerCountRangeTo int
}

// Common Cacher testing function type(s).
type testOperationFuncOriginal func(id int, c db.Cacher, K string, V interface{}, wgTest *sync.WaitGroup, ch chan interface{})
type testOperationFunc func(wO *workerOptions)

type workerOptions struct {
	// The implemented and initialized Cacher interface form the Cacher bank.
	Cache db.Cacher

	// A channel spinned off with the worker itself, should be closed by the worker (via the testOperationFunc).
	Chan chan interface{}

	// Identification number or the worker index.
	ID int

	// A stable key.
	Key string

	// Name of such testing operation being assigned to such worker.
	TestOpName string

	// Value for of from the cache.
	Val interface{}

	// Pointer to the WaitGroup issued by the TestHandler
	WGTest *sync.WaitGroup
}

//
//  DEFINITIONS
//

const REPORT_TEMPLATE = "%2d: [%s.%s(): %t]: \t K: %s, \t V: %v\n"

var defaultTestOperationNames = []string{
	"Load",
	"Store",
	"Delete",
}

// Set of testing functions to evaluate varios Cacher interface implementations.
var defaultTestOperations = []testOperationFunc{
	// Cacher.Load() method test.
	func(wO *workerOptions) {
		defer wO.WGTest.Done()

		// Load the value from such c cache.
		V, loaded := wO.Cache.Load(wO.Key)

		// Compose and send a report.
		if wO.Chan != nil {
			wO.Chan <- fmt.Sprintf(REPORT_TEMPLATE, wO.ID, wO.Cache.GetName(), wO.TestOpName, loaded, wO.Key, V)
			close(wO.Chan)
		}
	},
	// Cacher.Store() method test.
	func(wO *workerOptions) {
		defer wO.WGTest.Done()

		// Try to assert the given interface's type. Otherwise go generate a new value.
		V, ok := wO.Val.(int)
		if !ok {
			V = rand.IntN(100)
		}

		// Store the random value as key.
		stored := wO.Cache.Store(wO.Key, V)

		// Compose and send a report.
		if wO.Chan != nil {
			wO.Chan <- fmt.Sprintf(REPORT_TEMPLATE, wO.ID, wO.Cache.GetName(), wO.TestOpName, stored, wO.Key, V)
			close(wO.Chan)
		}
	},
	// Cacher.Delete() method test.
	func(wO *workerOptions) {
		defer wO.WGTest.Done()

		// Delete the value associated with such K key.
		deleted := wO.Cache.Delete(wO.Key)

		var V interface{} = nil

		// Compose and send a report.
		if wO.Chan != nil {
			wO.Chan <- fmt.Sprintf(REPORT_TEMPLATE, wO.ID, wO.Cache.GetName(), wO.TestOpName, deleted, wO.Key, V)
			close(wO.Chan)
		}
	},
}

// A very simple key generator lambda.
var GenerateKeys = func(N, L int) []string {
	if N == 0 || L == 0 {
		return []string{}
	}

	var keys []string

	for i := 0; i < N; i++ {
		keys = append(keys, String(L))
	}

	return keys
}

//
//  MAIN LAMBDAS
//

// timeoutHandler is a special lambda for to combine time.Duration (resp. time.Sleep func) and sync.WaitGroup Wait method.
// This lambda should block the main thread.
var TimeoutHandler = func(timeChan chan bool, timeout time.Duration, wgTest *sync.WaitGroup) {
	if timeChan == nil {
		panic("--- Cannot write to the time channel: nil")
	}

	if timeChan != nil {
		defer close(timeChan)
	}

	// Start the timeout goroutine.
	go func(ch chan bool) {
		time.Sleep(timeout)

		if ch != nil {
			fmt.Printf("--- Timeout reached! Printing results received so far...\n\n")
			ch <- true
		}
	}(timeChan)

	// Start the WaitGroup goroutine.
	go func(ch chan bool) {
		wgTest.Wait()

		if ch != nil {
			ch <- true
		}
	}(timeChan)

	// Wait for the first message.
	<-timeChan
}

// TestHandler is a relatively complex lambda containing the main worker group settings, and their execution. This can run in goroutine safely.
// <wgMain> pointer is passed from the main() function.
var TestHandler = func(tC *testConfiguration, wgMain *sync.WaitGroup) {
	if wgMain == nil {
		return
	}
	defer wgMain.Done()

	var wgTest sync.WaitGroup

	// Fetch the number of workers.
	N := tC.WorkerCount

	// Choose a random number too if is specified so.
	if tC.WorkerCountRandom {
		N = rand.IntN(tC.WorkerCountRangeTo) + 1
	}

	// Prepare the key set.
	tC.Keys = GenerateKeys(tC.KeyCount, tC.KeyLen)

	fmt.Printf("Starting %d workers (timeout set to %d ms)...\n\n", N, tC.Timeout/1000000)
	wgTest.Add(N)

	// Spinoff new concurrent workers.
	for idx := 0; idx < N; idx++ {
		// Create a channel for such worker.
		tC.TestChannels = append(tC.TestChannels, make(chan interface{}, 1))

		// New random value for the cache.
		V := rand.IntN(1000)

		// Fetch the random cache, operation and key number.
		cacheNo := rand.IntN(len(tC.CacheBank))
		opNo := rand.IntN(len(tC.TestOperations))
		keyNo := rand.IntN(len(tC.Keys))

		wO := &workerOptions{
			Cache:      tC.CacheBank[cacheNo],
			Chan:       tC.TestChannels[idx],
			ID:         idx,
			Key:        tC.Keys[keyNo],
			TestOpName: tC.TestOperationNames[opNo],
			Val:        V,
			WGTest:     &wgTest,
		}

		// Execute the run.
		go tC.TestOperations[opNo](wO)
	}

	// Wait for all workers to see the final results afterwards.
	var timeChan = make(chan bool, 1)
	TimeoutHandler(timeChan, tC.Timeout, &wgTest)
}

// ReadTestResults is a relatively simple lambda to fetch and print all the test results.
var ResultHandler = func(tC *testConfiguration) {
	// Fetch the common fan-in channels channel.
	chh := db.FanInChannels(nil, tC.TestChannels...)

	// Fetch the result interface{} and assert its type..
	for R := range chh {
		res, ok := R.(string)
		if !ok {
			fmt.Printf("--- Invalid result...\n")
		}

		// Print the report.
		fmt.Printf("%s\n", res)
	}
}

//
//  AutOINIT
//

var defaultTestConfiguration *testConfiguration

func init() {
	var defaultCacheBank []db.Cacher

	// Various db.Cacher interface implementations could be appended here.
	defaultCacheBank = append(defaultCacheBank, db.NewDefaultCache("default"))
	defaultCacheBank = append(defaultCacheBank, db.NewSimpleCache("simple"))
	defaultCacheBank = append(defaultCacheBank, db.NewSignalCache("signal"))

	// Return the pointer to the test configuration.
	defaultTestConfiguration = &testConfiguration{
		CacheBank:          defaultCacheBank,
		Keys:               []string{},
		KeyLen:             5,
		KeyCount:           3,
		TestChannels:       make([]chan interface{}, 0),
		TestOperations:     defaultTestOperations,
		TestOperationNames: defaultTestOperationNames,
		Timeout:            7500 * time.Millisecond,
		WorkerCount:        0,
		WorkerCountRandom:  true,
		WorkerCountRangeTo: 20,
	}
}
