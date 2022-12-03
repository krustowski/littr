package main

func main() {
	// Backend
	initBackend()

	// Frontend
	// https://github.com/maxence-charriere/go-app/issues/627
	initWASM()
	initServer()
}
