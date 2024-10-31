package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"go.vxn.dev/littr/pkg/backend/db"
)

var cacheBank = []db.Cacher{}

func init() {
	cacheBank = append(cacheBank, db.NewDefaultCache("default"))
	cacheBank = append(cacheBank, db.NewSimpleCache("simple"))
	cacheBank = append(cacheBank, db.NewSignalCache("signal"))
}

type opsFunc func(id int, c db.Cacher, K string, V interface{}, wg *sync.WaitGroup, ch chan interface{})

// Cacher operations.
var operations = []opsFunc{
	// Cacher.Load()
	func(id int, c db.Cacher, K string, V interface{}, wgTest *sync.WaitGroup, ch chan interface{}) {
		defer wgTest.Done()

		var loaded bool

		// Load the value from such c cache.
		V, loaded = c.Load(K)

		// Compose and send a report.
		if ch != nil {
			ch <- fmt.Sprintf("%2d: [%s.Load(): %t]: \t K: %s, \t V: %v\n", id, c.GetName(), loaded, K, V)
			close(ch)
		}
	},
	// Cacher.Store()
	func(id int, c db.Cacher, K string, V interface{}, wgTest *sync.WaitGroup, ch chan interface{}) {
		defer wgTest.Done()

		//V := Dummy{Name: K, Number: rand.IntN(100), Float: rand.FloatN(100)}
		//V = Dummy{Name: K, Number: rand.IntN(100)}
		V = rand.IntN(100)

		// Store the random value as key.
		stored := c.Store(K, V)

		// Compose and send a report.
		if ch != nil {
			ch <- fmt.Sprintf("%2d: [%s.Store(): %t]: \t K: %s, \t V: %v\n", id, c.GetName(), stored, K, V)
			close(ch)
		}
	},
	// Cacher.Delete()
	func(id int, c db.Cacher, K string, V interface{}, wgTest *sync.WaitGroup, ch chan interface{}) {
		defer wgTest.Done()

		// Delete the value associated with such K key.
		deleted := c.Delete(K)

		// Compose and send a report.
		if ch != nil {
			ch <- fmt.Sprintf("%2d: [%s.Delete(): %t]: \t K: %s, \t V: %v\n", id, c.GetName(), deleted, K, V)
			close(ch)
		}
	},
}

var testChans = make([]chan interface{}, 0)

var keys = func(N int) []string {
	var kk []string

	for i := 0; i < N; i++ {
		kk = append(kk, String(5))
	}

	return kk
}(2)

// runSimpleTest is a complex lambda containing the main worker group settings, and their execution. This can run in goroutine safely.
var runSimpleTest = func(wgMain *sync.WaitGroup) {
	defer wgMain.Done()

	var wgTest sync.WaitGroup

	// Fetch the number of workers.
	N := rand.IntN(15) + 1

	fmt.Printf("Starting %d workers...\n\n", N)
	wgTest.Add(N)

	// Spinoff new concurrent workers.
	for i := 0; i < N; i++ {
		// Create a channel for such worker.
		testChans = append(testChans, make(chan interface{}, 1))

		// Generate new key and value.
		//key := RandStringBytesMaskImprSrcUnsafe(5)
		//key := String(5)
		num := rand.IntN(1000)

		// Fetch the random cache, operation and key number.
		cacheNo := rand.IntN(len(cacheBank))
		opNo := rand.IntN(len(operations))
		keyNo := rand.IntN(len(keys))

		// Execute the run.
		go operations[opNo](
			i,
			cacheBank[cacheNo],
			keys[keyNo],
			num,
			&wgTest,
			testChans[i],
		)
	}

	// Wait for all workers to see the final results afterwards.
	//wgTest.Wait()
	var timeChan = make(chan bool, 1)
	timeoutHandler(timeChan, timeoutDur, &wgTest)
}

// readTestResults is a relatively simple lambda to fetch and print all the test results.
var readTestResults = func() {
	// Fetch the fan-in channel
	chh := db.FanInChannels(nil, testChans...)

	// Fetch the result interface{}
	for R := range chh {
		res, ok := R.(string)
		if !ok {
			fmt.Printf("--- Invalid result...\n")
		}

		fmt.Printf("%s\n", res)
	}
}

type timeoutChan chan bool

var timeoutDur = 7500 * time.Millisecond

// timeoutHandler is a special lambda for to combine time.Duration (resp. time.Sleep func) and sync.WaitGroup Wait method.
// This lambda should block the main thread.
var timeoutHandler = func(timeChan chan bool, timeout time.Duration, wg *sync.WaitGroup) {
	if timeChan != nil {
		defer close(timeChan)
	}

	// Start the timeout goroutine.
	go func(ch *chan bool) {
		time.Sleep(timeout)

		if ch != nil && *ch != nil {
			fmt.Printf("--- Timeout reached! Printing results received so far...\n\n")
			*ch <- true
		}
	}(&timeChan)

	// Start the waitgroup goroutine.
	go func(ch *chan bool) {
		wg.Wait()

		if ch != nil && *ch != nil {
			*ch <- true
		}
	}(&timeChan)

	// Wait for the first message.
	<-timeChan
}

//
//  MAIN
//

func main() {
	var wgMain sync.WaitGroup

	wgMain.Add(1)

	// Set the workers execution, and start them.
	go runSimpleTest(&wgMain)

	// Wait for all workers to finish (or hit the deadline). Whatever comes the first, it unblock the main thread.
	wgMain.Wait()
	//var timeChan = make(chan bool, 1)
	//timeoutHandler(timeChan, timeoutDur, &wgMain)

	// Fetch and print the test results.
	readTestResults()

	return
}
