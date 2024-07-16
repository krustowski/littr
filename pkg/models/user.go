package models

import (
	"time"
)

type User struct {
	// ID is an unique identifier.
	ID string `json:"id" binding:"required" validation:"required"`

	// Nickname is a login name of such user.
	Nickname string `json:"nickname" binding:"required"`

	// FullName is the "genuine" name of such user.
	FullName string `json:"full_name"`

	// Passphrase is a hashed pass phrase string (binary form).
	Passphrase string `json:"passphrase,omitempty"`

	// PassphraseHex is a hashed pass phrase string (hexadecimal alphanumberic form).
	PassphraseHex string `json:"passphrase_hex,omitempty"`

	// Email is a primary user's e-mail address.
	Email string `json:"email,omitempty"`

	// Web is user's personal homepage.
	Web string `json:"web"`

	// AvatarURL is an URL to the user's custom profile picture.
	AvatarURL string `json:"avatar_url,omitempty"`

	// About is a description string of such user.
	About string `json:"about"`

	// Active boolean indicates an activated user's account.
	Active bool `json:"active"`

	// Private boolean indicates a private user's account.
	Private bool `json:"private"`

	// FlowList is a string map of users, which posts should be added to one's flow page.
	FlowList map[string]bool `json:"flow_list,omitempty"`

	// ShadeList is a map of account/users to be shaded (soft-blocked) from following.
	ShadeList map[string]bool `json:"shade_list,omitempty"`

	// RequestList is a map of account requested to add this user to their flow --- used with the Private property.
	RequestList map[string]bool `json:"request_list,omitempty"`

	// FlowToggle is a single implementation of FlowList.
	FlowToggle string `json:"flow_toggle"`

	// Color is the user's UI color scheme.
	Color string `json:"color" default:"#000000"`

	// RegisteredTime is an UNIX timestamp of the user's registration.
	RegisteredTime time.Time `json:"registered_time"`

	// LastLoginTime is an UNIX timestamp of the last user's successful log-in.
	LastLoginTime time.Time `json:"last_login_time"`

	// LastLoginTime is an UNIX timestamp of the last action performed by such user.
	LastActiveTime time.Time `json:"last_active_time"`

	// searched is a bool indicating a status for the search engine.
	Searched bool `json:"-" default:true`

	// GDPR consent, set to true because it is noted on the registration page so. No user data should
	// be saved if the boolean is false.
	GDPR bool `json:"gdpr" default:true`

	// AppBgMode string defines the colour mode of the app's background (light vs dark).
	AppBgMode string `json:"app_bg_mode" default:"dark"`

	// Tags is an array of possible roles and other various attributes assigned to such user.
	Tags []string `json:"tags"`
}

// UserStat is a helper struct to hold statistics about the whole app.
type UserStat struct {
	// PostCount is a number of posts of such user.
	PostCount int `default:0`

	// ReactionCount tells the number of interactions (stars given).
	ReactionCount int `default:0`

	// FlowerCount is basically a number of followers.
	FlowerCount int `default:0`

	// ShadeCount is basically a number of blockers.
	ShadeCount int `default:0`

	// Searched is a special boolean used by the search engine to mark who is to be shown in search results.
	Searched bool `default:true`
}
