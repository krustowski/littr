package models

import (
	"time"
)

type Request struct {
	// Unique UUID.
	ID string `json:"id"`

	// User's name to easily fetch user's data from the database.
	Nickname string `json:"nickname"`

	// Requesting user's e-mail address.
	Email string `json:"email"`

	// Timestamp of the request generation, should expire in 24 hours after creation.
	CreatedAt time.Time `json:"created_at"`

	// Type is a helper field to differentiate the request's processor target.
	Type string `json:"type"`
}

func (r Request) GetID() string {
	return r.ID
}
