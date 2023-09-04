package models

import "time"

type Log struct {
	ID        int       `json:"id"`
	Nickname  string    `json:"nickname"`
	IP        string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
}
