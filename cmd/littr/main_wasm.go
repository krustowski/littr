//go:build wasm

package main

func main() {
	c := newClient()
	c.Run()
}
