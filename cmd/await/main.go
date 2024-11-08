package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"go.vxn.dev/littr/pkg/backend/live"

	"github.com/go-chi/chi/v5"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type MyApp struct {
	app.Compo
	Data    string
	dataSSE string
}

// Pregenerated by ChatGPT.
func (m *MyApp) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("HTTP Fetch Example in Go using WASM"),
		app.Button().Text("Fetch Data").OnClick(m.onFetchData),
		app.If(m.Data != "", app.P().Text(m.Data)),
		app.Button().Text("SSE Listen").OnClick(m.onListenToSSE),
		app.If(m.dataSSE != "", app.P().Text(m.dataSSE)),
	)
}

// RequestInit is an options struct for the Fetch API usage.
// https://developer.mozilla.org/en-US/docs/Web/API/RequestInit
type RequestInit struct {
	Body           string            `json:"body"`
	Cache          string            `json:"cache" default:"default"`
	Credentials    string            `json:"credentials" default:"same-origin"`
	Headers        map[string]string `json:"headers"`
	Keepalive      bool              `json:"keepalive" default:"false"`
	Method         string            `json:"method" default:"GET"`
	Mode           string            `json:"mode" default:"cors"`
	Priority       string            `json:"priority" default:"auto"`
	Redirect       string            `json:"redirect" default:"follow"`
	Referrer       string            `json:"referrer" default:"about:client"`
	ReferrerPolicy string            `json:"referrerPolicy"`
	Signal         interface{}       `json:"signal"`
}

// Usable map for export to JSValue via app.ValueOf(x)
var DefaultRequestInit = map[string]interface{}{
	"body":           nil,
	"cache":          "default",
	"credentials":    "same-origin",
	"headers":        map[string]interface{}{},
	"keepalive":      "false",
	"method":         "GET",
	"mode":           "cors",
	"priority":       "auto",
	"redirect":       "follow",
	"referrer":       "about:client",
	"referrerPolicy": "",
	"signal":         nil,
}

func NewRequestInitOptions() *RequestInit {
	return &RequestInit{}
}

func (m *MyApp) onListenToSSE(ctx app.Context, e app.Event) {
	// Create a fetch request to read the stream
	promise := app.Window().Call("fetch", "http://localhost:8081/api/v1/live", map[string]interface{}{
		"cache":     "no-cache",
		"keepalive": true,
		"headers": map[string]interface{}{
			"Accept": "text/event-stream",
		},
	})

	// Handle the Promise result using a callback
	then := app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		response := args[0]
		reader := response.Get("body").Call("getReader")

		// Define a function to recursively read chunks
		var readChunk app.Value

		readChunk = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			readPromise := reader.Call("read")
			readPromise.Then(func(chunk app.Value) {
				// Get `done` and `value` from the resolved promise
				done := chunk.Get("done").Bool()
				value := chunk.Get("value")

				if done {
					fmt.Println("stream closed.")
					return
				}

				// Process the chunk
				decoder := app.Window().Get("TextDecoder").New("utf-8")
				text := decoder.Call("decode", value)
				fmt.Println("Received data:", text.String())

				/*ctx.Dispatch(func(ctx app.Context) {
					m.dataSSE = text.String()
				})*/

				// Continue reading the next chunk
				readChunk.Invoke()
				return
			})
			return nil
		})

		// Start the first read
		readChunk.Invoke()
		return nil
	})

	// Start the fetch and handle it in the `then` callback
	promise.Call("then", then)
}

func (m *MyApp) onFetchData(ctx app.Context, e app.Event) {
	go func() {
		url := "http://localhost:8081/api/v1/live"

		//
		//  options --- Go map
		//

		goOpts := DefaultRequestInit
		goOpts["cache"] = "no-cache"
		goOpts["headers"] = map[string]interface{}{"Accept": "text/event-stream"}
		goOpts["keepalive"] = true
		goOpts["method"] = "GET"

		//
		//  options --- JSValue
		//

		signal := app.Window().Get("AbortSignal").Call("timeout", app.ValueOf(10000))
		jsOpts := app.ValueOf(goOpts)
		jsOpts.Set("signal", signal)

		//
		//  Fetch
		//

		promise := app.Window().Call("fetch", url, jsOpts)
		promise.Then(func(response app.Value) {
			if response.Get("status").Int() != 200 {
				// Notify the UI's component.
				ctx.Dispatch(func(ctx app.Context) {
					m.Data = handleStatus(response)
				})
				return
			}

			//
			//  JSON response handling
			//

			if response.Get("ok").Bool() {
				// [object Response]
				//fmt.Printf("%s\n", response.Call("toString").String())

				// {}
				//fmt.Printf("%s\n", app.Window().Get("JSON").Call("stringify", response).String())

				subpromise := response.Call("json")
				subpromise.Then(func(result app.Value) {
					// [object Object]
					//fmt.Printf("%s\n", result.Call("toString").String())

					// {"message": "lmao"}
					//fmt.Printf("%s\n", app.Window().Get("JSON").Call("stringify", result).String())
					rawJSON := app.Window().Get("JSON").Call("stringify", result)
					//rawJSON.Call("catch", app.Window().Get("catchError"))

					var resultMap map[string]interface{}

					// Unmarshal the raw string to []byte stream JSON to the interface map.
					if err := json.Unmarshal([]byte(rawJSON.String()), &resultMap); err != nil {
						fmt.Printf("%s\n", err.Error())
					}

					// Update the UI's component.
					ctx.Dispatch(func(ctx app.Context) {
						m.Data = rawJSON.String()
					})

					// Assert type string.
					msg, ok := resultMap["message"].(string)
					if !ok {
						return
					}

					// Logs "msg: lmao" to console.
					fmt.Printf("msg: %s\n", msg)
				})
				subpromise.Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
					ctx.Dispatch(func(ctx app.Context) {
						m.Data = this.Get("result").String()
					})

					return args[0].Get("message").String()
				}))
			}
		})

		catch := app.Window().Get("catchError")

		promise.Call("catch", catch)
	}()
}

// Handle unexpected HTTP Status code
var handleStatus = func(response app.Value) string {
	if !response.IsNull() {
		return fmt.Sprintf("Unexpected HTTP status code: %d (%s).\n", response.Get("status").Int(), response.Get("statusText").String())
	}
	return ""
}

// Handle fetch errors
var catchError = app.FuncOf(func(this app.Value, args []app.Value) interface{} {
	//defer catchError.Release()
	err := args[0].Get("message").String()

	//app.Window().Get("console").Call("log", app.ValueOf("test"))
	app.Window().Get("console").Call("log", args[0])

	// Log to console.
	fmt.Printf("Fetch error: %s\n", err)

	// Update the UI's component.
	/*ctx.Dispatch(func(ctx app.Context) {
		m.Data = err
	})*/

	return err
})

// Export a Go function to JavaScript to interact with the DOM (for WASM purposes)
func init() {
	app.Window().Set("catchError", catchError)

	app.Window().Set("fetchData", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		go func() {
			fmt.Println("fetchData function called from JavaScript")
		}()
		return nil
	}))
}

//
//  dummy HTTP server
//

func main() {
	defer catchError.Release()

	app.Route("/", &MyApp{})
	app.RunWhenOnBrowser()

	r := chi.NewRouter()

	// Create a custom network TCP connection listener.
	listener, err := net.Listen("tcp", "0.0.0.0:8081")
	if err != nil {
		// Cannot listen on such address = a permission issue?
		panic(err)
	}
	defer listener.Close()

	server := &http.Server{
		Addr:         listener.Addr().String(),
		WriteTimeout: 0 * time.Second,
		Handler:      r,
	}

	// Build up a simple "backend API service".
	r.Get("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		jData, err := json.Marshal(&struct {
			Message string `json:"message"`
		}{
			Message: "lmao",
		})
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(jData)
	})

	r.Post("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var data = struct {
			Test string `json:"test"`
		}{}

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
			return
		}

		err = json.Unmarshal(reqBody, &data)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
			return
		}
		w.Write([]byte(fmt.Sprintf("{\"lmaoooo\": \"%s\"}", data.Test)))

		/*jData, err := json.Marshal(&struct {
			Message string `json:"message"`
		}{
			Message: data.Test,
		})
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(jData)*/
	})

	r.Mount("/api/v1/live", live.Streamer)

	go func() {
		for {
			if live.Streamer == nil {
				fmt.Printf("Streamer ded...\n")
				break
			}

			live.BroadcastMessage(live.EventPayload{Type: "keepalive", Data: "heartbeat"})

			time.Sleep(3 * time.Second)
		}
	}()

	r.Mount("/", &app.Handler{})

	if err := server.Serve(listener); err != nil {
		fmt.Errorf(err.Error())
	}
}
