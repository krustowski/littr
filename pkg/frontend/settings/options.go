package settings

import (
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/models"
)

type optionsPayload struct {
	UIMode        bool         `json:"ui_mode"`
	UITheme       models.Theme `json:"ui_theme"`
	LiveMode      bool         `json:"live_mode"`
	LocalTimeMode bool         `json:"local_time_mode"`
	Private       bool         `json:"private"`
	AboutText     string       `json:"about_you"`
	WebsiteLink   string       `json:"website_link"`
}

func (c *Content) prefillPayload() optionsPayload {
	payload := optionsPayload{
		UIMode:        c.user.UIMode,
		UITheme:       c.user.UITheme,
		LiveMode:      c.user.Options["liveMode"],
		LocalTimeMode: c.user.Options["localTimeMode"],
		Private:       c.user.Options["private"],
		AboutText:     c.user.About,
		WebsiteLink:   c.user.Web,
	}

	return payload
}

func (c *Content) updateOptions(payload optionsPayload) {
	if payload.UIMode != c.user.UIMode {
		body := app.Window().Get("document").Call("querySelector", "body")
		currentClass := body.Get("className").String()
		parts := strings.Split(currentClass, "-")

		mode := func() string {
			if len(parts) != 2 {
				return "light-orang"
			}

			if parts[0] == "light" {
				return "dark-" + parts[1]
			}

			return "light-" + parts[1]
		}()

		body.Set("className", mode)
	}

	if payload.UITheme != c.user.UITheme {
		body := app.Window().Get("document").Call("querySelector", "body")
		currentClass := body.Get("className").String()
		parts := strings.Split(currentClass, "-")

		// This is going to be replaced with switch soon.
		theme := func() string {
			if len(parts) != 2 {
				return "dark-orang"
			}

			if parts[1] == "blu" {
				return parts[0] + "-orang"
			}

			return parts[0] + "-blu"
		}()

		body.Set("className", theme)
	}

	c.user.UIMode = payload.UIMode
	c.user.UITheme = payload.UITheme
	c.user.Options["liveMode"] = payload.LiveMode
	c.user.Options["localTimeMode"] = payload.LocalTimeMode
	c.user.Options["private"] = payload.Private
	c.user.About = payload.AboutText
	c.user.Web = payload.WebsiteLink
}
