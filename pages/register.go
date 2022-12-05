package pages

import (
	"litter-go/backend"
	"net/mail"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type RegisterPage struct {
	app.Compo
}

type registerContent struct {
	app.Compo

	toastShow bool
	toastText string

	nickname   string
	passphrase string
	email      string
}

func (p *RegisterPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("register / littr")
}

func (p *RegisterPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&registerContent{},
	)
}

func (c *registerContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.nickname != "" && c.passphrase != "" && c.email != "" {
			// validate e-mail struct
			// https://stackoverflow.com/a/66624104
			if _, err := mail.ParseAddress(c.email); err != nil {
				c.toastText = "wrong e-mail format entered"
				c.toastShow = true
				return
			}

			if ok := backend.AddUser(c.nickname, c.passphrase, c.email); ok {
				c.toastShow = false
				ctx.Navigate("/login")
				return
			}
			c.toastShow = true
			c.toastText = "backend error!"
		}
	})
}

func (c *registerContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *registerContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr registration"),
		app.P().Text("do not be mid, join us to be lit"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field label border deep-orange-border deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true),
			app.Label().Text("nickname"),
		),
		app.Div().Class("field label border deep-orange-border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border deep-orange-border deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.email)).Required(true),
			app.Label().Text("e-mail"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("register").OnClick(c.onClick),
	)
}
