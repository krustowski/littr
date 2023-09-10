package models

import (
	"time"
)

var (
	LogsChan chan Log
)

// Log struct describes the format of a logging service's JSON output.
type Log struct {
	// Nickname is the currently logged user's name.
	Nickname string `json:"nickname" default:"guest"`

	// IPAddress is the remote user's IP address string, of type IPv4/IPv6.
	IPAddress string `json:"ip_address"`

	// Timestamp is an UNIX timestamp of the logged action.
	Timestamp time.Time `json:"timestamp"`

	// Caller describes the logger engine caller (e.g. failed auth, new post etc)
	Caller string `json:"caller_name"`

	// Content describes the log's very content, message etc.
	Message string `json:"content"`
}

func NewLog(ipAddress string) *Log {
	log := Log{
		Timestamp: time.Now(),
	}
	return &log
}

func (l *Log) Content(message string) {
	l.Message = message
	LogsChan <- *l
	return
}
