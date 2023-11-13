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

	// Passphrase is a hashed pass phrase string.
	Passphrase string `json:"passphrase"`

	// Email is a primary user's e-mail address.
	Email string `json:"email"`

	// Web is user's personal homepage.
	Web string `json:"web"`

	// AvatarURL is an URL to the user's custom profile picture.
	AvatarURL string `json:"avatar_url"`

	// About is a description string of such user.
	About string `json:"about"`

	// Active boolean indicates an activated user's account.
	Active bool `json:"active"`

	// FlowList is a string map of users, which posts should be added to one's flow page.
	FlowList map[string]bool `json:"flow_list"`

	// ShadeList is a map of account/users to be shaded (soft-blocked) from following.
	ShadeList map[string]bool `json:"shade_list"`

	// FlowToggle is a single implementation of FlowList.
	FlowToggle string `json:"flow_toggle"`

	// Color is the user's UI color scheme.
	Color string `json:"color" default:"#000000"`

	// RegisteredTime is an UNIX timestamp of the user's registeration.
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

	// ReplyNotificationOn is a bool, that indicates the state of notification permission made by user. 
	// Is set to false (off) on default.
	ReplyNotificationOn bool `json:"reply_notification_on" default:false`

	// VapidPubKey is a string containing VAPID public key for notification subscription.
	VapidPubKey string `json:"vapid_pubkey"`

	// VapidPrivKey is a string containing VAPID private key for notification subscription.
	VapidPrivKey string `json:"vapid_privkey"`
}
