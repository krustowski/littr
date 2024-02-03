package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Logger struct {
	// CallerID is a nickname of the user calling the API.
	CallerID string `json:"caller_id" validation:"required"`

	// Code integer is a HTTP return code.
	Code int `json:"code" validation:"required"`

	// IPAddress string is basically an user's IPv4/IPv6 address (beware of proxies).
	IPAddress string `json:"ip_address"`

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
}

func NewLogger(r *http.Request, worker string) *Logger {
	if r == nil || worker == "" {
		return nil
	}

	return &Logger{
		CallerID:   r.Header.Get("X-API-Caller-ID"),
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
