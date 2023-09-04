package backend

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

type Post struct {
	// Unique post ID.
	ID int `json:"id"`

	// Post type --- post, poll, reply, img
	Type string `json:"type"`

	// Author's account name.
	Nickname string `json:"nickname"`

	// Base64 encoded content.
	Content string `json:"content"`

	// UNIX timestamp of the post publication.
	Timestamp time.Time `json:"timestamp"`

	// Poll content.
	Poll Poll `json:"poll"`

	// Post ID being replied to.
	ReplyTo int `json:"reply_to"`

	// ReactionCount counts the number of item's reactions.
	ReactionCount int `json:"reaction_count"`
}

type Log struct {
	ID        int    `json:"id"`
	Nickname  string `json:"nickname"`
	IP        string `json:"ip_address"`
	Timestamp int    `json:"timestamp"`
	Action    string `json:"action"`
}

type Global struct{}

type Poll struct {
	ID          int        `json:"id"`
	Question    string     `json:"question"`
	OptionOne   PollOption `json:"option_one"`
	OptionTwo   PollOption `json:"option_two"`
	OptionThree PollOption `json:"option_three"`
	Voted       []string   `json:"voted_list"`
	Timestamp   time.Time  `json:"timestamp"`
}

type PollOption struct {
	Content string `json:"content"`
	Counter int    `json:"counter"`
}

type Registration struct {
	ID         int    `json:""`
	Nickname   string `json:"nickname"`
	Passphrase string `json:"passphrase"`
	Poll       int    `json:"poll_id"`
}
