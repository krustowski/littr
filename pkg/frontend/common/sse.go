package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.vxn.dev/littr/pkg/config"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

const JS_LITTR_SSE = "littrServiceSSE"
const JS_LITTR_EVENT = "littrEventSSE"

//
//  Service options for fetch().
//

var fetchOpts = app.ValueOf(map[string]interface{}{
	"cache":     "no-cache",
	"keepalive": true,
	"headers": map[string]interface{}{
		"Accept": "text/event-stream",
	},
	"signal": nil,
})

var reconnTimeout int = 30000

//
//  Service methods.
//

var connect = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	// Channel of errors that is used with the main fetch() function.
	var chE chan string
	var err string

	defer func() {
		if chE != nil {
			close(chE)
		}
	}()

	if (!this.Get("running").IsUndefined() && this.Get("running").Bool()) || app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
		fmt.Printf("The service is already running")
		return "ServiceAlreadyRunningError"
	}

	//fmt.Printf("starting %s\n", this.Get("name").String())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// The signals monitoring goroutine.
	go func() {
		// Wait for signals.
		sig := <-sigs
		signal.Stop(sigs)

		fmt.Println("Caught signal: ", sig)

		this.Call("abort")
		app.Window().Get(JS_LITTR_SSE).Set("running", false)
	}()

	// Run the main fetch() logic, and look for errors.
	go FetchSSE(chE)

	go func() {
		err = <-chE
		if err != "OK" {
			return
		}
	}()

	return err
})

var reconnect = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	timeout := app.Window().Get(JS_LITTR_SSE).Get("reconnTimeout").Int()
	if timeout <= 0 {
		return nil
	}

	if app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
		return nil
	}

	var retryCount = config.MAX_SSE_RETRY_COUNT

	go func() {
		// Execute the new fetch() request. Set the timeout to 30000 milliseconds per retry.
		app.Window().Get(JS_LITTR_SSE).Set("signal", app.Window().Get("AbortSignal").Call("timeout", reconnTimeout))

		for {
			time.Sleep(time.Millisecond * time.Duration(timeout))

			if app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
				break
			}

			err := app.Window().Get(JS_LITTR_SSE).Call("connect")

			if retryCount == 0 {
				fmt.Println("No retries left fo reconnection")
			}

			// Check for the possible error returned by the function call.
			if !err.JSValue().IsNull() && retryCount > 0 {
				// Decrease the retry count, and invoke the call again.
				retryCount--
				continue
			}

			app.Window().Get(JS_LITTR_SSE).Set("running", true)
			break
		}
	}()

	return nil
})

var stop = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	fmt.Println("Stopping littr SSE client")

	if !this.Get("running").IsUndefined() {
		this.Set("running", false)
	}

	this.Call("tryReconnect")
	return nil
})

var abort = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	fmt.Println("Aborting the fetch controller")

	this.Set("running", false)
	this.Get("controller").Call("abort")

	this.Call("tryReconnect")
	return nil
})

// Export a Go function to JavaScript to interact with the DOM (for WASM purposes)
func init() {
	//
	//  SSE Service
	//

	// Export littrServiceSSE object.
	if app.Window().Get(JS_LITTR_SSE).IsUndefined() {
		app.Window().Set(JS_LITTR_SSE, map[string]interface{}{
			"name":          "littr SSE client",
			"fetchOpts":     map[string]interface{}{},
			"controller":    nil,
			"running":       false,
			"reconnTimeout": 15000,
			// Methods
			"connect":      nil,
			"stop":         nil,
			"abort":        nil,
			"tryReconnect": nil,
		})
	}

	var aController = app.Window().Get("AbortController").New()
	app.Window().Get(JS_LITTR_SSE).Set("controller", aController)
	fetchOpts.Set("signal", aController.Get("signal"))

	app.Window().Get(JS_LITTR_SSE).Set("fetchOpts", fetchOpts)

	app.Window().Get(JS_LITTR_SSE).Set("connect", connect)
	app.Window().Get(JS_LITTR_SSE).Set("stop", stop)
	app.Window().Get(JS_LITTR_SSE).Set("abort", abort)
	app.Window().Get(JS_LITTR_SSE).Set("tryReconnect", reconnect)

	fmt.Println("Connecting to the SSE stream...")
	app.Window().Get(JS_LITTR_SSE).Call("tryReconnect")

	//
	//  SSE Event
	//

	// Export littrEventSSE object.
	if app.Window().Get(JS_LITTR_EVENT).IsUndefined() {
		app.Window().Set(JS_LITTR_EVENT, map[string]interface{}{
			"id":        "",
			"type":      "",
			"data":      "",
			"translate": nil,
		})
	}
}

// FetchSSE is an early implementation of the SSE client using only await fetch() as the base function.
func FetchSSE(ch chan string) {
	defer func() {
		if ch != nil {
			close(ch)
		}
	}()

	// Check if the service isn't already running. If so, exit.
	if app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
		return //fmt.Sprintf("ServiceAlreadyRunningError")
	}

	// Mark the service as running.
	app.Window().Get(JS_LITTR_SSE).Set("running", true)

	// Create a fetch request to read the stream.
	promise := app.Window().Call("fetch", "/api/v1/live", app.Window().Get(JS_LITTR_SSE).Get("fetchOpts"))

	// Handle the Promise result using a callback.
	promise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		response := args[0]
		reader := response.Get("body").Call("getReader")

		if response.Get("status").Int() != 200 {
			app.Window().Get(JS_LITTR_SSE).Call("abort")

			return fmt.Sprintf(response.Get("statusText").String())
		}

		fmt.Println("Connected")

		// Define a function to recursively read chunks.
		var readChunk app.Value

		readChunk = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			readPromise := reader.Call("read")
			readPromise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				// Get `done` and `value` from the resolved promise.
				chunk := args[0]
				done := chunk.Get("done").Bool()
				value := chunk.Get("value")

				if done {
					fmt.Println("Stream closed")
					if ch != nil {
						ch <- fmt.Sprintf("StreamClosedError")
					}
					app.Window().Get(JS_LITTR_SSE).Call("abort")
					return nil
				}

				if ch != nil {
					ch <- fmt.Sprintf("OK")
				}

				// Process the chunk into a SSE event.
				decoder := app.Window().Get("TextDecoder").New("utf-8")
				text := decoder.Call("decode", value).String()

				event := NewSSEEvent(text)
				fmt.Printf("%s\n", event.Dump())

				// Create a new HTML DOM event.
				domE := app.Window().Get("document").Call("createEvent", "HTMLEvents")
				domE.Call("initEvent", "message", true, true)
				domE.Set("eventName", event.Type)
				domE.Set("data", event.Data)

				// Send the HTML event (handled by eventListeners).
				app.Window().Call("dispatchEvent", domE)

				// Continue reading the next chunk.
				if app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
					readChunk.Invoke()
				}

				return nil

			})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				err := args[0].Get("name").String()
				fmt.Println("Body reader error caught: ", err)

				if ch != nil {
					ch <- fmt.Sprintf(err)
				}

				app.Window().Get(JS_LITTR_SSE).Call("stop")
				return nil
			}))

			return nil
		})

		// Start the first read.
		readChunk.Invoke()
		return nil

	})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		err := args[0].Get("name").String()

		if err == "AbortError" {
		}

		if ch != nil {
			ch <- fmt.Sprintf(err)
		}

		fmt.Println("Fetch error caught: ", err)
		app.Window().Get(JS_LITTR_SSE).Call("stop")
		return nil
	}))

	return
}
