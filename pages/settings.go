package pages

import (
	"litter-go/backend"

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

	toastShow bool
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

func (c *settingsContent) onClickPass(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		c.toastShow = true
		if c.passphrase == "" || c.passphraseAgain == "" {
			c.toastText = "both passphrases need to be filled"
			return
		}

		if c.passphrase != c.passphraseAgain {
			c.toastText = "passphrases do not match"
			return
		}

		if ok := backend.EditUserPassword(c.passphrase); !ok {
			c.toastText = "generic backend error"
			return
		}

		c.toastShow = false
		ctx.Navigate("/settings")
	})
}

func (c *settingsContent) onClickAbout(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		c.toastShow = true
		if c.aboutText == "" {
			c.toastText = "about textarea needs to be filled"
			return
		}

		if ok := backend.EditUserAbout(c.aboutText); !ok {
			c.toastText = "generic backend error"
			return
		}

		c.toastShow = false
		ctx.Navigate("/settings")
	})
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *settingsContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings"),
		app.P().Text("change your passphrase, the about string or just fuck off ;)"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)),
			app.Label().Text("passphrase again"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClickPass),

		app.Div().Class("large-divider"),

		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Text("change me").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout),
	)
}
