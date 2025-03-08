package settings

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type OptionsPayload struct {
	UIMode        bool   `json:"ui_mode"`
	LiveMode      bool   `json:"live_mode"`
	LocalTimeMode bool   `json:"local_time_mode"`
	Private       bool   `json:"private"`
	AboutText     string `json:"about_you"`
	WebsiteLink   string `json:"website_link"`
}

func (c *Content) prefillPayload() OptionsPayload {
	payload := OptionsPayload{
		UIMode:        c.user.Options["uiMode"],
		LiveMode:      c.user.Options["liveMode"],
		LocalTimeMode: c.user.Options["localTimeMode"],
		Private:       c.user.Options["private"],
		AboutText:     c.user.About,
		WebsiteLink:   c.user.Web,
	}

	return payload
}

func (c *Content) updateOptions(payload OptionsPayload) {
	if payload.UIMode != c.user.Options["uiMode"] {
		/*ctx.LocalStorage().Set("mode", "dark")
		if !c.darkModeOn {
			ctx.LocalStorage().Set("mode", "light")
		}*/

		app.Window().Get("LIT").Call("toggleMode")
	}

	c.user.Options["uiMode"] = payload.UIMode
	c.user.Options["liveMode"] = payload.LiveMode
	c.user.Options["localTimeMode"] = payload.LocalTimeMode
	c.user.Options["private"] = payload.Private
	c.user.About = payload.AboutText
	c.user.Web = payload.WebsiteLink
}
