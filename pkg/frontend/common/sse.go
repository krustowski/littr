package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

const JS_LITTR_SSE = "littrServiceSSE"
const JS_LITTR_EVENT = "littrEventSSE"

//
//  Service options for fetch()
//

var opts = app.ValueOf(map[string]interface{}{
	"cache":     "no-cache",
	"keepalive": true,
	"headers": map[string]interface{}{
		"Accept": "text/event-stream",
	},
	"signal": nil,
})

//
//  Service methods
//

	var run = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		if (!this.Get("running").IsUndefined() && this.Get("running").Bool()) || app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
			fmt.Printf("the service is already running")
			return nil
		}

		fmt.Printf("starting %s\n", this.Get("name").String())

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		// The signals monitoring goroutine.
		go func() {
			// Wait for signals.
			<-sigs
			signal.Stop(sigs)

			this.Call("abort")
		}()

		/*go func() {
			time.Sleep(time.Second * 20)

			this.Call("abort")
		}()*/

		go FetchSSE(nil)

		return nil
	})

	var stop = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		fmt.Printf("stopping littr SSE client")
		if !this.Get("running").IsUndefined() {
			this.Set("running", false)
		}
		return nil
	})

	var abort = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		fmt.Printf("aborting the fetch controller")
		this.Get("controller").Call("abort")
		app.Window().Get(JS_LITTR_SSE).Set("running", true)
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
			"name":       "littr SSE client",
			"opts":       map[string]interface{}{},
			"controller": nil,
			"running":    false,
			// Methods
			"run":   nil,
			"stop":  nil,
			"abort": nil,
		})
	}

	var aController = app.Window().Get("AbortController").New()
	app.Window().Get(JS_LITTR_SSE).Set("controller", aController)
	opts.Set("signal", aController.Get("signal"))

	//var timeout = app.Window().Get("AbortSignal").Call("timeout", 25000)
	//opts.Set("signal", timeout)

	app.Window().Get(JS_LITTR_SSE).Set("opts", opts)

	app.Window().Get(JS_LITTR_SSE).Set("run", run)
	app.Window().Get(JS_LITTR_SSE).Set("stop", stop)
	app.Window().Get(JS_LITTR_SSE).Set("abort", abort)

	app.Window().Get(JS_LITTR_SSE).Call("run")

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
func FetchSSE() {
	defer func() {
		if ch != nil {
			close(ch)
		}
	}()

	var chT chan time.Time

	if app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
		return
	}

	app.Window().Get(JS_LITTR_SSE).Set("running", true)

	// Create a fetch request to read the stream
	promise := app.Window().Call("fetch", "/api/v1/live", app.Window().Get(JS_LITTR_SSE).Get("opts"))

	// Handle the Promise result using a callback
	promise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		response := args[0]
		reader := response.Get("body").Call("getReader")

		// Define a function to recursively read chunks
		var readChunk app.Value

		readChunk = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			readPromise := reader.Call("read")
			readPromise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				// Get `done` and `value` from the resolved promise
				chunk := args[0]
				done := chunk.Get("done").Bool()
				value := chunk.Get("value")

				if done {
					fmt.Println("stream closed")
					if ch != nil {
						close(ch)
					}
					return nil
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

				app.Window().Call("dispatchEvent", domE)

				// Pass the received event to the channel.
				if ch != nil {
					ch <- *event
				}

				// Continue reading the next chunk.
				if app.Window().Get(JS_LITTR_SSE).Get("running").Bool() {
					readChunk.Invoke()
				}

				return nil
			})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				//err := args[0]

				fmt.Println(args[0].Get("name").String())
				fmt.Println(args[0].Get("message").String())
				return nil
			}))

			return nil
		})

		// Start the first read.
		readChunk.Invoke()
		return nil

	})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		fmt.Printf(args[0].Get("message").String())
		fmt.Printf("uuuuhhhhhhh")
		return nil
	}))

	return
}
