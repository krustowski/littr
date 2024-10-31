package main

import (
	"sync"
)

//
//  MAIN
//

func main() {
	// Should be initialized automatically.
	if defaultTestConfiguration == nil {
		return
	}

	var wgMain sync.WaitGroup

	// Set the workers execution, and start them.
	wgMain.Add(1)
	go TestHandler(defaultTestConfiguration, &wgMain)

	// Wait for all workers to finish (or hit the deadline).
	wgMain.Wait()

	// Fetch and print the test results.
	ResultHandler(defaultTestConfiguration)

	return
}
