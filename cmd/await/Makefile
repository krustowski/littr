.PHONY: fmt
fmt:
	@gofmt -w -s .

.PHONY: run
run: fmt
	@echo -e "Building await demo..."
	@GOOS=js GOARCH=wasm go build -o web/app.wasm main.go
	@echo -e "Running await demo..."
	@go run main.go

