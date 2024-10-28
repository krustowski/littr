//go:build wasm
// +build wasm

package main

func initClient() {
	initClientCommon()
}

func initServer() {
	// function initServer() is blanked here to reduce the final WASM binary file size, which is used on the client's side (see build at the top of this file)
}
