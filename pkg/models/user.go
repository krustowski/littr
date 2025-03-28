package models

import (
	"time"
)

type UserOptionsMap map[string]bool

// DefaultUserOptionsMap can be assigned directly at the register time to a new user, or to be used in the Options migrations as a template.
var DefaultUserOptionsMap = UserOptionsMap{
	"active":        false,
	"gdpr":          true,
	"private":       false,
	"uiMode":        true,
	"liveMode":      true,
	"localTimeMode": true,
}

// UserGenericMap is used for user lists mainly at the moment.
type UserGenericMap map[string]bool

type User struct {
	// Nickname is a login name of such user.
	Nickname string `json:"nickname" binding:"required" example:"alice"`

	// FullName is the "genuine" name of such user.
	FullName string `json:"full_name"`

	// Passphrase is a hashed pass phrase string (binary form).
	Passphrase string `json:"passphrase,omitempty" swaggerignore:"true"`

	// PassphraseHex is a hashed pass phrase string (hexadecimal alphanumberic form).
	PassphraseHex string `json:"passphrase_hex,omitempty" swaggerignore:"true"`

	// Email is a primary user's e-mail address.
	Email string `json:"email,omitempty" example:"alice@example.com"`

	// Web is user's personal homepage.
	Web string `json:"web" example:"https://example.com"`

	// AvatarURL is an URL to the user's custom profile picture.
	AvatarURL string `json:"avatar_url,omitempty" example:"https://example.com/web/apple-touch-icon.png"`

	// About is a description string of such user.
	About string `json:"about" default:"newbie"`

	// Options is an umbrella struct/map for the booleans.
	Options UserOptionsMap `json:"options" example:"private:true"`

	// Active boolean indicates an activated user's account.
	Active bool `json:"active" example:"true"`

	// Private boolean indicates a private user's account.
	Private bool `json:"private" example:"true"`

	// FlowList is a string map of users, which posts should be added to one's flow page.
	FlowList UserGenericMap `json:"flow_list,omitempty" example:"alice:true"`

	// ShadeList is a map of account/users to be shaded (soft-blocked) from following.
	ShadeList UserGenericMap `json:"shade_list,omitempty" example:"cody:true"`

	// RequestList is a map of account requested to add this user to their flow --- used with the Private property.
	RequestList UserGenericMap `json:"request_list,omitempty" example:"dave:true"`

	// Color is the user's UI color scheme.
	Color string `json:"color" default:"#000000"`

	// RegisteredTime is an UNIX timestamp of the user's registration.
	RegisteredTime time.Time `json:"registered_time"`

	// LastLoginTime is an UNIX timestamp of the last user's successful log-in.
	LastLoginTime time.Time `json:"last_login_time"`

	// LastLoginTime is an UNIX timestamp of the last action performed by such user.
	LastActiveTime time.Time `json:"last_active_time"`

	// searched is a bool indicating a status for the search engine.
	Searched bool `json:"-" swaggerignore:"true"`

	// GDPR consent, set to true because it is noted on the registration page so. No user data should
	// be saved if the boolean is false.
	GDPR bool `json:"gdpr"`

	// UIMode bool defines the colour mode of the app's background (light vs dark).
	UIMode bool `json:"ui_mode"`

	UITheme Theme `json:"ui_theme"`

	// LiveMode is a feature allowing to show notifications about new posts
	LiveMode bool `json:"live_mode"`

	// LocalTimeMode is a feature to show any post's datetime in the local time according to the client's/user's device setting.
	LocalTimeMode bool `json:"local_time_mode"`

	// Devices array holds the subscribed devices. Devices are not exported as the subscribed devices are stored separated.
	Devices []Device `json:"devices" swaggerignore:"true"`

	// Tags is an array of possible roles and other various attributes assigned to such user.
	Tags []string `json:"tags" example:"user"`
}

func (u User) Copy() *User {
	return &u
}

func (u User) GetID() string {
	return u.Nickname
}

// Options is an umbrella struct to hold all the booleans in one place.
type Options struct {
	// Active boolean indicates an activated user's account.
	// Map equivalent: active
	Active bool `json:"active" default:"true"`

	// GDPR consent, set to true because it is noted on the registration page so. No user data should
	// be saved if the boolean is false.
	// Map equivalent: gdpr
	GDPR bool `json:"gdpr" default:"true"`

	// Private boolean indicates a private user's account.
	// Map equivalent: private
	Private bool `json:"private" default:"false"`

	// AppBgMode string defines the colour mode of the app's background (light vs dark).
	// Map equivalent: uiMode
	UIDarkMode bool `json:"app_bg_mode" default:"true"`

	// LiveMode is a feature allowing to show notifications about new posts
	// Map equivalent: liveMode
	LiveMode bool `json:"live_mode" default:"true"`

	// LocalTimeMode is a feature to show any post's datetime in the local time according to the client's/user's device setting.
	// Map equivalent: localTimeMode
	LocalTimeMode bool `json:"local_time_mode" default:"true"`
}

// UserStat is a helper struct to hold statistics about the whole app.
type UserStat struct {
	// PostCount is a number of posts of such user.
	PostCount int64 `json:"post_count" default:"0"`

	// ReactionCount tells the number of interactions (stars given).
	ReactionCount int64 `json:"reaction_count" default:"0"`

	// FlowerCount is basically a number of followers.
	FlowerCount int64 `json:"flower_count" default:"0"`

	// ShadeCount is basically a number of blockers.
	ShadeCount int64 `json:"shade_count" default:"0"`

	// Searched is a special boolean used by the search engine to mark who is to be shown in search results.
	Searched bool `json:"searched" default:"true" swaggerignore:"true"`
}

type Theme int

const (
	ThemeDefault Theme = iota
	ThemeOrang
)

func (t Theme) Bg() string {
	var labels = []string{
		"blue",
		"deep-orange7",
	}

	return labels[t]
}

func (t Theme) Text() string {
	var labels = []string{
		"blue-text",
		"deep-orange-text",
	}

	return labels[t]
}

func (t Theme) Border() string {
	var labels = []string{
		"blue-border",
		"deep-orange-border",
	}

	return labels[t]
}
