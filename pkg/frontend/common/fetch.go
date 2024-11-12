package common

import (
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

// RequestInit object. Usable map for export to JSValue via app.ValueOf(x)
var DefaultRequestInit = map[string]interface{}{
	// Serialized JSON string.
	"body": nil,
	// Caching options: default, no-store, reload, no-cache, force-cache, only-if-cached.
	"cache": "default",
	// Credential control options: omit, same-site (def.), include.
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
	"url":            "",
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

	// Convert body into a string.
	if input.Data != nil && init["method"] != "GET" && init["method"] != "HEAD" {
		jsonData, err := json.Marshal(input.Data)
		if err != nil {
			return false
		}

		init["body"] = string(jsonData)
	} else {
		init["body"] = nil
	}

	// Set the abort controller signal callback.
	/*var aController = app.Window().Get("AbortController").New()
	app.Window().Set("fetchController", aController)
	init["signal"] = aController.Get("signal")*/

	// Call the ligher fetch wrapper.
	out, code, err := Fetch(&init)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	// Read again and associate fields.
	r := strings.NewReader(*out)

	// Unmarshal the raw string to []byte stream JSON to the interface map.
	if err := json.NewDecoder(r).Decode(&output); err != nil {
		fmt.Println(err.Error())
		return false
	}

	// Assign the HTTP response code.
	output.Code = code

	return true
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

	// The initial fetch call with options to get the promise.
	promise := app.Window().Call("fetch", (*input)["url"], *input)
	promise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		response := args[0]

		//
		//  Response handling
		//
		if response.Get("status").Int() != 200 {
			chC <- response.Get("status").Int()
			chE <- fmt.Errorf("unexpected HTTP status code: %d (%s)", response.Get("status").Int(), response.Get("statusText").String())
			chS <- ""

			close(chC)
			close(chE)
			close(chS)

			//app.Window().Get("fetchController").Call("abort")
			return nil
		}

		//
		//  JSON response handling
		//
		if response.Get("ok").Bool() {
			// Another promise getter via the JSON function call upon the response object.
			// --> fetch(url).then(response => response.json())
			subpromise := response.Call("json")
			subpromise.Call("then", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				result := args[0]

				// Preprocess the response to be unserializable. And send to output.
				chC <- 200
				chS <- app.Window().Get("JSON").Call("stringify", result).String()
				chE <- nil

				close(chC)
				close(chE)
				close(chS)

				//app.Window().Get("fetchController").Call("abort")
				return nil

			})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
				chC <- 500
				chE <- fmt.Errorf("%s\n", args[0].Get("message").String())
				chS <- ""

				close(chC)
				close(chE)
				close(chS)

				//app.Window().Get("fetchController").Call("abort")
				return nil
			}))
		}
		return nil

		// Error catching callback for the main fetch() promise.
	})).Call("catch", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		chE <- fmt.Errorf("%s", args[0].Get("message").String())
		chS <- ""

		close(chC)
		close(chE)
		close(chS)

		return nil
	}))

	// Catch the results.
	output := <-chS
	code := <-chC
	err := <-chE

	//app.Window().Get("fetchController").Call("abort")

	return &output, code, err
}

// Handle unexpected HTTP Status code
var handleStatus = func(response app.Value) string {
	if !response.IsNull() {
		return fmt.Sprintf("Unexpected HTTP status code: %d (%s).\n", response.Get("status").Int(), response.Get("statusText").String())
	}
	return ""
}
