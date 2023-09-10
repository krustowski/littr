package models

import (
	"time"
)

type Session struct {
	// ID is a session hashed ID.
	ID string `json:"id"`

	// Nickname is the user's name.
	Nickname string `json:"nickname"`

	// LastActionTime is an UNIX timestamp of the last action requested by such user.
	// Should be revoked after 24 hours?
	LastActionTime time.Time `json:"last_action_time"`

	// CreatedAtTime is an UNIX timestamp of the session's creation.
	CreatedAtTime time.Time `json:"created_at_time"`

	// Active bool indicates whether is such action still active.
	Active bool `json:"active"`

	// IPAdress is a string describing used IPv4-type address by such action.
	// Should be destroyed when logged from other IP address.
	// TODO: catch IPv6 too.
	IPAddress string `json:"ip_address"`
}
