package settings

import (
// "github.com/maxence-charriere/go-app/v10/pkg/app"
)

type OptionsPayload struct {
	UIDarkMode    bool   `json:"dark_mode"`
	LiveMode      bool   `json:"live_mode"`
	LocalTimeMode bool   `json:"local_time_mode"`
	Private       bool   `json:"private"`
	AboutText     string `json:"about_you"`
	WebsiteLink   string `json:"website_link"`
}

func (c *Content) prefillPayload() OptionsPayload {

	payload := OptionsPayload{
		UIDarkMode:    c.user.UIDarkMode,
		LiveMode:      c.user.LiveMode,
		LocalTimeMode: c.user.LocalTimeMode,
		Private:       c.user.Private,
		AboutText:     c.user.About,
		WebsiteLink:   c.user.Web,
	}

	return payload
}
