package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Logger interface {
	// Basic methods: setters.
	Msg(message string) Logger
	Status(code int) Logger
	Error(err error) Logger
	SetPrefix(prefix string) Logger
	RemovePrefix() Logger

	// Getters.
	CallerID() string

	// Timer-related methods.
	ResetTimer() Logger

	// Data-related methods.
	Log() Logger
	Payload(pl interface{}) Logger
	Write(w http.ResponseWriter)
}

const (
	// Localhost as the IPv4 address.
	LOCALHOST4 = "127.0.0.1"

	// Localhost as the IPv6 address.
	LOCALHOST6 = "::1"
)

type DefaultLogger struct {
	// CallerID is a nickname of the user calling the API.
	callerID string `json:"-"`

	// Code integer is a HTTP return code.
	Code int `json:"code" validation:"required"`

	// IPAddress string is basically an user's IPv4/IPv6 address (beware of proxies).
	IPAddress string `json:"-"`

	// Message string holds a custom message returned by a various HTTP handler.
	Message string `json:"message" validation:"required"`

	// Prefix stands at the start of the logger output before the very message.
	Prefix string `json:"-"`

	// Method string hold a HTTP method name.
	Method string `json:"method"`

	// Route string is the very route called by user.
	Route string `json:"route"`

	// Version is the tagged version of the client's SW (compiled in).
	Version string `json:"version"`

	// WorkerName string is the name of a worker processing such request.
	WorkerName string `json:"worker_name" validation:"required"`

	// Response is a helper field to hold the prepared API response for sending.
	Response *APIResponse `json:"-"`

	// Err is a helper error field to hold the error type from the BE logging callback procedure.
	Err error `json:"-"`

	// TimerStart holds the starting point of the time measurement. To be subtracted and written to the JSON output afterwards.
	TimerStart time.Time `json:"request_start"`

	// TimeStop is the very stop time mark in terms of the system/application process duration.
	TimerStop time.Time `json:"request_stop"`

	// TimeDurationNS hold the difference between the start and stop time points regarding the application process making its duration (in nanoseconds).
	TimerDurationNS time.Duration `json:"request_duration_ns"`
}

// NewLogger takes the HTTP request structure in (can be nil), and the worker's name (required string) to prepare a new Logger instance.
// Returns a pointer to the new Logger instance.
func NewLogger(r *http.Request, worker string) Logger {
	// Worker name has to be set always.
	if worker == "" {
		return nil
	}

	// Start the Timer just now.
	start := time.Now()

	// Decide the version's name.
	version := func() string {
		if r != nil && r.Header.Get("X-App-Version") != "" {
			return r.Header.Get("X-App-Version")
		}

		return "system"
	}()

	// Little hotfix for the data dump/load procedure.
	if r == nil {
		return &DefaultLogger{
			callerID:   "system",
			IPAddress:  LOCALHOST4,
			Method:     "",
			Message:    "",
			Prefix:     "",
			Route:      "",
			TimerStart: start,
			Version:    version,
			WorkerName: worker,
		}
	}

	// Fetch the caller's nickname, to be checked if not blank afterwards.
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok || callerID == "" {
		callerID = "system"
	}

	return &DefaultLogger{
		callerID:   callerID,
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Message:    "",
		Prefix:     "",
		Route:      r.URL.String(),
		TimerStart: start,
		Version:    version,
		WorkerName: worker,
	}
}

// encode works as a simple macro returning JSON-encoded string of the Logger struct.
func (l *DefaultLogger) encode() string {
	jsonString, err := json.Marshal(l)
	if err != nil {
		fmt.Println("error marshalling Logger struct (", err.Error(), ")")
		return err.Error()
	}

	return string(jsonString[:])
}

// CallerID returns the caller's ID (nickname), that has been (hopefully) decided from the context.
func (l DefaultLogger) CallerID() string {
	return l.callerID
}

// ResetTimer resets the TimerStart timestamp. Usable in the procedures where the logger is passed (???)
// or not to log the whole HTTP server uptime (the gracefully HTTP server shutdown goroutine).
func (l DefaultLogger) ResetTimer() Logger {
	l.TimerStart = time.Now()
	return l
}

//
//  Prefix-related methods
//

// SetPrefix sets the log's prefix according to the input <prefix> string.
func (l DefaultLogger) SetPrefix(prefix string) Logger {
	l.Prefix = prefix
	return l
}

// RemovePrefix remove preiously prepended string from the Logger struct.
func (l DefaultLogger) RemovePrefix() Logger {
	l.Prefix = ""
	return l
}

//
//  Basic methods
//

// Msg writes the input <msg> string to the Logger struct for its following output.
func (l DefaultLogger) Msg(msg string) Logger {
	var message string

	if l.Prefix != "" {
		message = l.Prefix + ": "
	}

	l.Message = message + msg
	return l
}

// Status writes the HTTP Status code (as integer) for the following logger output.
func (l DefaultLogger) Status(code int) Logger {
	l.Code = code
	return l
}

// Error takes an error and holds it in the Logger structure for the possible output.
func (l DefaultLogger) Error(err error) Logger {
	l.Err = err
	return l
}

//
//  Data-related methods
//

// Log write the logger's JSON output to the stdout.
func (l DefaultLogger) Log() Logger {
	if l.IPAddress == "" {
		l.IPAddress = LOCALHOST4
	}

	if l.Err != nil {
		l.Message += " (" + l.Err.Error() + ")"
	}

	// Stop the count!
	l.TimerStop = time.Now()
	l.TimerDurationNS = l.TimerStop.Sub(l.TimerStart)

	fmt.Println(l.encode())
	return l
}

// Payload takes and prepares the HTTP response's body payload. The input can be nil.
func (l DefaultLogger) Payload(pl interface{}) Logger {
	// construct the generic API response
	l.Response = &APIResponse{
		Message:   l.Message,
		Timestamp: time.Now().UnixNano(),
		Data:      pl,
	}

	return l
}

// Write writes the HTTP headers and sends the response to the client.
func (l DefaultLogger) Write(w http.ResponseWriter) {
	jsonData, err := json.Marshal(l.Response)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Default HTTP code with no body.
	if l.Response == nil && jsonData == nil {
		l.Code = http.StatusNoContent
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(l.Code)

	if _, err := io.Writer.Write(w, jsonData); err != nil {
		fmt.Println(err.Error())
	}
}

//
//  DummyLogger is the dummy implementation of the LoggerInterface. It suits as a blank logger just above the level of a nil.
//

type DummyLogger struct{}

func NewDummyLogger() *DummyLogger {
	return &DummyLogger{}
}

func (l DummyLogger) Msg(message string) Logger {
	return l
}

func (l DummyLogger) Status(code int) Logger {
	return l
}

func (l DummyLogger) Error(_ error) Logger {
	return l
}

func (l DummyLogger) SetPrefix(prefix string) Logger {
	return l
}

func (l DummyLogger) RemovePrefix() Logger {
	return l
}

func (l DummyLogger) CallerID() string {
	return ""
}

func (l DummyLogger) ResetTimer() Logger {
	return l
}

func (l DummyLogger) Log() Logger {
	return l
}

func (l DummyLogger) Payload(pl interface{}) Logger {
	return l
}

func (l DummyLogger) Write(w http.ResponseWriter) {}
