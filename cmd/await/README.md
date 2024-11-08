# await fetch() implementation using go-app's JS wrapper

This is just an attempt on how to make the WASM client binary even more lighter with the gzip compression.

The implementation also handles errors via the `catch(fn)` function with a callback.

Main source of inspiration: https://github.com/maxence-charriere/go-app/issues/995#issuecomment-2394535140


## how to use

In shell run these:

```
cd cmd/await
mkdir -p web

GOOS=js GOARCH=wasm go build -o web/app.wasm main.go
go run main.go
```

Then:

+ open your web browser and navigate to [http://localhost:8081/](http://localhost:8081/)
+ press F12 to see the console log
+ press the button in the page to see the effect
