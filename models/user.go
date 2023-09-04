package models

import "time"

type User struct {
	// Unique user ID.
	ID int `json:"id" binding:"required" validation:"required"`

	// User nickname.
	Nickname string `json:"nickname" binding:"required"`

	// Hashed user password.
	Passphrase string `json:"passphrase"`

	// User's personal e-mail.
	Email string `json:"email"`

	// Little user description.
	About string `json:"about"`

	// Important boolean to indicate user's active status; required for login.
	Active bool `json:"active"`

	// List of other accounts to show in the flow.
	FlowList []string `json:"flow_list"`

	// Field used for flow user add/removal.
	FlowToggle string `json:"flow_toggle"`

	// UI custom color.
	Color string `json:"color" default:"#000000"`

	// UNIX timestamp of the last login.
	LastLoginTime time.Time `json:"last_login"`

	// UNIX timestamp of the last UI interaction -- useful for 'show online'.
	LastActiveTime time.Time `json:"last_active"`
}
