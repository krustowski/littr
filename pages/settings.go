package pages

import (
	"litter-go/backend"
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type SettingsPage struct {
	app.Compo
}

type settingsContent struct {
	app.Compo

	passphrase      string
	passphraseAgain string
	aboutText       string

	showToast bool
	toastText string
}

func (p *SettingsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&settingsContent{},
	)
}

func (p *SettingsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("settings / littr")
}

func (c *settingsContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.passphrase != "" && c.passphraseAgain != "" {
			if c.passphrase == c.passphraseAgain {
				if ok := backend.EditUserPassword(c.passphrase); ok {
					log.Println(c.passphrase)
					c.showToast = false
					ctx.Navigate("/flow")
				}
			} else {
				c.showToast = true
			}
		}
	})
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	c.showToast = false
}

func (c *settingsContent) Render() app.UI {
	toastActiveClass := ""
	if c.showToast {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings"),
		app.P().Text("change your passphrase, the about string or just fuck off ;)"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text("passphrases do not match!"),
			),
		),

		/*
			app.Div().Class("modal"+toastActiveClass).Body(
				app.H5().Text("Default modal"),
				app.Div().Body(
					app.Text("Some text here"),
				),
				app.Nav().Class("right-align").Body(
					app.Button().Class("border").Text("Cancel"),
					app.Button().Text("Confirm"),
				),
			),
		*/

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)),
			app.Label().Text("passphrase again"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClick),
		app.Div().Class("large-divider"),

		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Text("idk").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about"),

		app.Div().Class("large-space"),
		app.Div().Class("large-space"),
	)
}
