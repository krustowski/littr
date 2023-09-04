package main

import (
	"litter-go/config"
)

func main() {
	config.Init()

	// https://github.com/maxence-charriere/go-app/issues/627
	initWASM()
	initServer()
}
