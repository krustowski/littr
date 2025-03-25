//go:build wasm

package main

func main() {
	var c = newClient()
	c.Run()
}
