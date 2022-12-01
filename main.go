package main

func main() {
	// Backend
	backendServe()

	// Frontend
	// https://github.com/maxence-charriere/go-app/issues/627
	initWASM()
	initServer()
}
