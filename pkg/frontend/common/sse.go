package common

import (
	"encoding/json"
	"fmt"
	"time"

	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

const (
	JsLittrSse   = "littrServiceSSE"
	JsLittrEvent = "littrEventSSE"
)

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

	if (!this.Get("running").IsUndefined() && this.Get("running").Bool()) || app.Window().Get(JsLittrSse).Get("running").Bool() {
		fmt.Printf("The service is already running")
		return "ServiceAlreadyRunningError"
	}

	/*sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// The signals monitoring goroutine.
	go func() {
		// Wait for signals.
		sig := <-sigs
		signal.Stop(sigs)

		fmt.Println("Caught signal: ", sig)

		this.Call("abort")
		app.Window().Get(JsLittrSse).Set("running", false)
	}()*/

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
	if app.Window().Get(JsLittrSse).Get("reconnRunning").Bool() {
		return nil
	}
	app.Window().Get(JsLittrSse).Set("reconnRunning", true)

	timeout := app.Window().Get(JsLittrSse).Get("reconnTimeout").Int()
	if timeout <= 0 {
		return nil
	}

	// Use the firstTimeout when the app is loaded for the first time. Invalide it right after its value usage.
	if app.Window().Get(JsLittrSse).Get("firstTimeout").Int() > 0 {
		timeout = app.Window().Get(JsLittrSse).Get("firstTimeout").Int()
		app.Window().Get(JsLittrSse).Set("firstTimeout", 0)
	}

	// No need to reconnect if the service is running.
	if app.Window().Get(JsLittrSse).Get("running").Bool() {
		return nil
	}

	var retryCount = config.MaxSseRetryCount

	go func() {
		// Execute the new fetch() request. Set the timeout to 30000 milliseconds per retry.
		app.Window().Get(JsLittrSse).Set("signal", app.Window().Get("AbortSignal").Call("timeout", reconnTimeout))

		for {
			time.Sleep(time.Millisecond * time.Duration(timeout))

			if app.Window().Get(JsLittrSse).Get("running").Bool() {
				break
			}

			err := app.Window().Get(JsLittrSse).Call("connect")

			if retryCount == 0 {
				fmt.Println("No retries left fo reconnection")
			}

			// Check for the possible error returned by the function call.
			if !err.JSValue().IsNull() && retryCount > 0 {
				// Decrease the retry count, and invoke the call again.
				retryCount--
				continue
			}

			app.Window().Get(JsLittrSse).Set("running", true)
			break
		}
	}()

	return nil
})

var stop = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	fmt.Println("Stopping littr SSE client")

	this.Set("running", false)

	this.Call("tryReconnect")
	return nil
})

var abort = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	fmt.Println("Aborting the fetch controller")

	this.Set("running", false)
	this.Get("controller").Call("abort")

	var ac = app.Window().Get("AbortController").New()
	this.Set("controller", ac)
	fetchOpts.Set("signal", ac.Get("signal"))

	this.Call("tryReconnect")
	return nil
})

// Export a Go function to JavaScript to interact with the DOM (for WASM purposes)
func init() {
	//
	//  SSE Service
	//

	// Export littrServiceSSE object.
	if app.Window().Get(JsLittrSse).IsUndefined() {
		app.Window().Set(JsLittrSse, map[string]interface{}{
			"name": "littr SSE client",
			// Options
			"fetchOpts":     map[string]interface{}{},
			"controller":    nil,
			"reconnTimeout": 15000,
			"firstTimeout":  2000,
			// Runtime booleans
			"running":       false,
			"reconnRunning": false,
			// Methods
			"connect":      nil,
			"stop":         nil,
			"abort":        nil,
			"tryReconnect": nil,
		})
	}

	// Set the abort controller signal callback.
	var aController = app.Window().Get("AbortController").New()
	app.Window().Get(JsLittrSse).Set("controller", aController)
	fetchOpts.Set("signal", aController.Get("signal"))

	// Set the options.
	app.Window().Get(JsLittrSse).Set("fetchOpts", fetchOpts)

	// Set the methods.
	app.Window().Get(JsLittrSse).Set("connect", connect)
	app.Window().Get(JsLittrSse).Set("stop", stop)
	app.Window().Get(JsLittrSse).Set("abort", abort)
	app.Window().Get(JsLittrSse).Set("tryReconnect", reconnect)
}

// FetchSSE is an early implementation of the SSE client using only await fetch() as the base function.
func FetchSSE(ch chan string) {
	defer func() {
		if ch != nil {
			close(ch)
		}
	}()

	// Check if the service isn't already running. If so, exit.
	if app.Window().Get(JsLittrSse).Get("running").Bool() {
		return //fmt.Sprintf("ServiceAlreadyRunningError")
	}

	// Mark the service as running.
	app.Window().Get(JsLittrSse).Set("running", true)

	// Create a fetch request to read the stream.
	promise := app.Window().Call("fetch", "/api/v1/live", app.Window().Get(JsLittrSse).Get("fetchOpts"))

	// Handle the Promise result using a callback.
	promise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		response := args[0]
		reader := response.Get("body").Call("getReader")

		if response.Get("status").Int() != 200 {
			app.Window().Get(JsLittrSse).Call("abort")

			return response.Get("statusText").String()
		}

		fmt.Println("Connected")
		app.Window().Get(JsLittrSse).Set("reconnRunning", false)

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
						ch <- "StreamClosedError"
					}
					app.Window().Get(JsLittrSse).Call("abort")
					return nil
				}

				if ch != nil {
					ch <- "OK"
				}

				// Process the chunk into a SSE event.
				decoder := app.Window().Get("TextDecoder").New("utf-8")
				text := decoder.Call("decode", value).String()

				event := NewSSEEvent(text)
				fmt.Printf("%s\n", text)
				//fmt.Printf("%s\n", event.Dump())

				// Create a new HTML DOM event.
				domE := app.Window().Get("document").Call("createEvent", "HTMLEvents")
				domE.Call("initEvent", "message", true, true)
				domE.Set("eventName", event.Type)
				domE.Set("data", event.Data)

				// Send the HTML event (handled by eventListeners).
				app.Window().Call("dispatchEvent", domE)

				// The last beat's timestamp save procedure.
				LS := app.Window().Get("localStorage")
				if !LS.IsNull() {
					LS.Call("setItem", "lastEventTime", time.Now().UnixNano())
				}

				var userStr string

				LS = app.Window().Get("localStorage")
				if !LS.IsNull() && !LS.Call("getItem", "user-data").IsUndefined() {
					userStr = LS.Call("getItem", "user-data").String()
				}

				userStruct := struct {
					Value models.User `json:"Value"`
				}{}

				// Unmarshal the result to get an User struct.
				err := json.Unmarshal([]byte(userStr), &userStruct)
				if err != nil {
					fmt.Println(err.Error())
					return nil
				}

				toastText, toastLink, keep := event.ParseEventData(&userStruct.Value)

				tPl := &ToastPayload{
					Name:  "snackbar-general-bottom",
					Text:  toastText,
					Link:  toastLink,
					Color: "blue10",
					Keep:  keep,
				}

				// Show the generic snackbar/toast.
				ShowGenericToast(tPl)

				// Continue reading the next chunk.
				if app.Window().Get(JsLittrSse).Get("running").Bool() {
					readChunk.Invoke()
				}

				return nil

				// Catch errors.
			})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				err := args[0].Get("name").String()
				fmt.Println("Body reader error caught: ", err)

				if err == "TypeError" {
					fmt.Println(args[0].Get("message").String())
				}

				if ch != nil {
					ch <- err
				}

				app.Window().Get(JsLittrSse).Call("stop")
				return nil
			}))

			return nil
		})

		// Start the first read.
		readChunk.Invoke()
		return nil

		// Catch errors.
	})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		err := args[0].Get("name").String()

		if ch != nil {
			ch <- err
		}

		fmt.Println("Fetch error caught: ", err)
		app.Window().Get(JsLittrSse).Call("stop")
		return nil
	}))
}
