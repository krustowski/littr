package common

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// DTO-in structure for the API call.
type CallInput struct {
	Method      string
	Url         string
	CallerID    string
	PageNo      int
	HideReplies bool

	// Payload body for the API call.
	Data interface{}
}

// Standardized common response from API.
type Response struct {
	Code      int
	Error     error
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`

	// Custom Data field per the app component. Must be referenced via the inner pointer too when the output is required.
	//
	//   Example:
	//   output := &common.Response{&model.User}
	//
	Data interface{} `json:"data"`
}

// Usable map for export to JSValue via app.ValueOf(x)
var DefaultRequestInit = map[string]interface{}{
	"body":           nil,
	"cache":          "default",
	"credentials":    "same-origin",
	"headers":        map[string]interface{}{},
	"keepalive":      false,
	"method":         "GET",
	"mode":           "cors",
	"priority":       "auto",
	"redirect":       "follow",
	"referrer":       "about:client",
	"referrerPolicy": "",
	"signal":         nil,
	// Other custom options.
	"callerID": "",
	"url":      "",
}

// FetchData is a metafunction for input options conversion for the main, lighter Fetch() function.
func FetchData(input *CallInput, output *Response) bool {
	init := DefaultRequestInit

	headers := map[string]interface{}{
		"X-Hide-Replies": input.HideReplies,
		"X-Page-No":      input.PageNo,
	}
	init["headers"] = headers

	if input.CallerID != "" {
		init["callerID"] = input.CallerID
	}

	if input.Method != "" {
		init["method"] = input.Method
	}

	init["url"] = input.Url

	if input.Data != nil && init["method"] != "GET" {
		jsonData, err := json.Marshal(input.Data)
		if err != nil {
			return false
		}

		init["body"] = string(jsonData)
	}

	// Call the ligher fetch wrapper.
	out, code, err := Fetch(&init)
	if err != nil {
		return false
	}

	// Read again and associate fields.
	r := strings.NewReader(*out)

	// Unmarshal the raw string to []byte stream JSON to the interface map.
	if err := json.NewDecoder(r).Decode(&output); err != nil {
		fmt.Println(err.Error())
		return false
	}

	output.Code = code

	return true
}

// FetchSSE is an early implementation of the SSE client using only await fetch() as the base function.
func FetchSSE(ch chan Event) {
	defer func() {
		if ch != nil {
			close(ch)
		}
	}()

	// Create a fetch request to read the stream
	promise := app.Window().Call("fetch", "/api/v1/live", map[string]interface{}{
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
					fmt.Println("stream closed")
					close(ch)
					return
				}

				// Process the chunk
				decoder := app.Window().Get("TextDecoder").New("utf-8")
				text := decoder.Call("decode", value)
				//fmt.Println("Received data:", text.String())

				// Read again and associate fields.
				r := strings.NewReader(text.String())
				b := bufio.NewReader(r)

				event := Event{}

				if len(strings.Split(text.String(), "\n")) >= 3 {
					for i := 0; i < 3; i++ {
						lineB, err := b.ReadSlice(byte('\n'))
						if err != nil {
							fmt.Println(err.Error())
							continue
						}

						//
						line := strings.Join(
							strings.Split(
								string(lineB), ":")[1:], " ")

						// Associate the line to event's fields.
						switch i {
						case 0:
							event.LastEventID = line
						case 1:
							event.Type = line
						case 2:
							event.Data = line
						}
					}
				}

				fmt.Printf("%s\n", event.Format())

				// Pass the received event to the channel.
				if ch != nil {
					ch <- event
				}

				// Continue reading the next chunk.
				readChunk.Invoke()
				return
			})
			return nil
		})

		// Start the first read.
		readChunk.Invoke()
		return nil
	})

	// Start the fetch and handle it in the `then` callback.
	promise.Call("then", then)

	return
}

// Fetch is a raw implementation of http.Client to omit the `net/*` packages completely. The main purpose is to further optimize the disk and memory space needed by the WASM app client.
func Fetch(input *map[string]interface{}) (*string, int, error) {
	if (*input)["url"] == "" {
		return nil, 0, fmt.Errorf("URL not specified for Fetch()")
	}

	// Start channels to catch the outputs.
	chC := make(chan int, 1)
	chE := make(chan error, 1)
	chS := make(chan string, 1)

	defer func() {
		close(chE)
		close(chS)
	}()

	// The initial fetch call with options to get the promise.
	promise := app.Window().Call("fetch", (*input)["url"], *input)
	promise.Then(func(response app.Value) {
		//
		//  Response handling
		//
		if response.Get("status").Int() != 200 {
			chC <- response.Get("status").Int()
			chE <- fmt.Errorf("unexpected HTTP status code: %d (%s)", response.Get("status").Int(), response.Get("statusText").String())
			chS <- ""
			return
		}

		//
		//  JSON response handling
		//
		if response.Get("ok").Bool() {
			// Another promise getter via the JSON function call upon the response object.
			// --> fetch(url).then(response => response.json())
			subpromise := response.Call("json")
			subpromise.Then(func(result app.Value) {
				// Preprocess the response to be unserializable. And send to output.
				chC <- 200
				chS <- app.Window().Get("JSON").Call("stringify", result).String()
				chE <- nil
				return
			})
			subpromise.Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				chC <- 500
				chE <- fmt.Errorf("%s\n", args[0].Get("message").String())
				chS <- ""
				return nil
			}))
		}
	})

	// Error catching callback for the main fetch() promise.
	promise.Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		chE <- fmt.Errorf("%s", args[0].Get("message").String())
		chS <- ""
		return nil
	}))

	// Catch the results.
	output := <-chS
	code := <-chC
	err := <-chE

	return &output, code, err
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
/*func init() {
	app.Window().Set("catchError", catchError)

	app.Window().Set("fetchData", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		go func() {
			fmt.Println("fetchData function called from JavaScript")
		}()
		return nil
	}))
}*/
