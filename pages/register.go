package pages

import (
	"litter-go/backend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type RegisterPage struct {
	app.Compo
}

type registerContent struct {
	app.Compo
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
		&registerContent{},
		&footer{},
	)
}

func (c *registerContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.nickname != "" && c.passphrase != "" && c.email != "" {
			//
			if ok := backend.AddUser(c.nickname, c.passphrase, c.email); ok {
				ctx.Navigate("/login")
			}
		}
	})
}

func (c *registerContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr registration"),
		app.P().Text("do not be mid, join us to be lit"),
		app.Div().Class("space"),

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
