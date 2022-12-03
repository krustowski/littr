package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type RegisterPage struct {
	app.Compo
}

type registerContent struct {
	app.Compo
}

func (p *SettingsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("register / littr")
}

func (p *RegisterPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&loginContent{},
		&footer{},
	)
}

func (c *registerContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr registration"),
		app.P().Text("do not be mid, join us to be lit"),

		app.Div().Class("field label border").Body(
			app.Input().Type("text"),
			app.Label().Text("nickname"),
		),
		app.Div().Class("field label border").Body(
			app.Input().Type("password"),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border").Body(
			app.Input().Type("text").Name("email"),
			app.Label().Text("e-mail"),
		),
		app.Button().Class("responsive primary").Text("register"),
	)
}
