package pages

import (
	"crypto/sha512"
	"litter-go/backend"
	"os"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type SettingsPage struct {
	app.Compo
}

type settingsContent struct {
	app.Compo

	loggedUser string

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

func (c *settingsContent) OnNav(ctx app.Context) {
	ctx.LocalStorage().Get("userName", &c.loggedUser)
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

		passHash := sha512.Sum512([]byte(c.passphrase + os.Getenv("APP_PEPPER")))

		if _, ok := litterAPI("PUT", "/api/users", backend.User{
			Nickname:   c.loggedUser,
			Passphrase: string(passHash[:]),
		}); !ok {
			c.toastText = "generic backend error"
			return
		}

		c.toastShow = false
		ctx.Navigate("/users")
	})
}

func (c *settingsContent) onClickAbout(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.aboutText == "" {
			c.toastShow = true
			c.toastText = "about textarea needs to be filled"
			return
		}

		if len(c.aboutText) > 100 {
			c.toastShow = true
			c.toastText = "about text has to be shorter than 100 chars"
			return
		}

		if _, ok := litterAPI("PUT", "/api/users", backend.User{
			Nickname: c.loggedUser,
			About:    c.aboutText,
		}); !ok {
			c.toastShow = true
			c.toastText = "generic backend error"
			return
		}

		c.toastShow = false
		ctx.Navigate("/users")
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

		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)),
			app.Label().Text("passphrase again"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClickPass),

		app.Div().Class("large-divider"),

		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Text("change me").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout),
	)
}
