//go:build !wasm

package main

import "runtime"

func main() {
	var c = newClient()
	c.Run()

	if runtime.GOOS != "wasm" {
		var s = newServer()
		s.Run()
	}
}
