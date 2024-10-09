package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Logger struct {
	// CallerID is a nickname of the user calling the API.
	CallerID string `json:"-"`

	// Code integer is a HTTP return code.
	Code int `json:"code" validation:"required"`

	// IPAddress string is basically an user's IPv4/IPv6 address (beware of proxies).
	IPAddress string `json:"-"`

	// Message string holds a custom message returned by a various HTTP handler.
	Message string `json:"message" validation:"required"`

	// Method string hold a HTTP method name.
	Method string `json:"method"`

	// Route string is the very route called by user.
	Route string `json:"route"`

	// Time property hold the actual time of the request processing.
	Time time.Time `json:"time" validation:"required"`

	// Version is the tagged version of the client's SW (compiled in).
	Version string `json:"version"`

	// WorkerName string is the name of a worker processing such request.
	WorkerName string `json:"worker_name" validation:"required"`

	Response *APIResponse
}

func NewLogger(r *http.Request, worker string) *Logger {
	if worker == "" {
		return nil
	}

	// little hack for data dump/load procedure
	if r == nil {
		return &Logger{
			CallerID:   "",
			IPAddress:  "127.0.0.1",
			Method:     "",
			Route:      "",
			WorkerName: worker,
			Version:    "",
		}
	}

	callerID, _ := r.Context().Value("nickname").(string)

	return &Logger{
		CallerID:   callerID,
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: worker,
		Version:    r.Header.Get("X-App-Version"),
	}
}

// encode works as a simple macro returning JSON-encoded string of the Logger struct.
func (l *Logger) encode() string {
	jsonString, err := json.Marshal(l)
	if err != nil {
		fmt.Println("error marshalling Logger struct -", err.Error())
		return err.Error()
	}

	return string(jsonString[:])
}

// Println formats the encoded Logger struct into an output string to stdin.
func (l *Logger) Println(msg string, code int) bool {
	l.Code = code
	l.Message = msg
	l.Time = time.Now()

	if l.IPAddress == "" {
		l.IPAddress = "127.0.0.1"
	}

	fmt.Println(l.encode())

	return true
}

// l.Msg("this user does not exist in the database").StatusCode(http.StatusNotFound).Log().Write(w)
func (l *Logger) Msg(msg string) *Logger {
	l.Message = msg
	return l
}

func (l *Logger) Status(code int) *Logger {
	l.Code = code
	return l
}

func (l *Logger) Log() *Logger {
	l.Time = time.Now()

	if l.IPAddress == "" {
		l.IPAddress = "127.0.0.1"
	}

	fmt.Println(l.encode())

	return l
}
func (l *Logger) Payload(pl interface{}) *Logger {
	if pl == nil {
		return l
	}

	// construct the generic API response
	l.Response = &APIResponse{
		Message:   l.Message,
		Timestamp: time.Now().UnixNano(),
		Data:      pl,
	}

	return l
}

func (l *Logger) Write(w http.ResponseWriter) {
	jsonData, err := json.Marshal(l.Response)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(l.Code)

	io.WriteString(w, fmt.Sprintf("%s", jsonData))
}
