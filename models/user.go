package models

import (
	"time"
)

type User struct {
	// ID is an unique identifier.
	ID string `json:"id" binding:"required" validation:"required"`

	// Nickname is a login name of such user.
	Nickname string `json:"nickname" binding:"required"`

	// Passphrase is a hashed pass phrase string.
	Passphrase string `json:"passphrase"`

	// Email is a primary user's e-mail address.
	Email string `json:"email"`

	// About is a description string of such user.
	About string `json:"about"`

	// Active boolean indicates an activated user's account.
	Active bool `json:"active"`

	// FlowList is a string array of users, which posts should be added to one's flow page.
	FlowList map[string]bool `json:"flow_list"`

	// FlowToggle is a single implementation of FlowList.
	FlowToggle string `json:"flow_toggle"`

	// Color is the user's UI color scheme.
	Color string `json:"color" default:"#000000"`

	// LastLoginTime is an UNIX timestamp of the last user's successful log-in.
	LastLoginTime time.Time `json:"last_login_time"`

	// LastLoginTime is an UNIX timestamp of the last action performed by such user.
	LastActiveTime time.Time `json:"last_active_time"`
}
