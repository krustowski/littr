package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type LoginPage struct {
	app.Compo
	userLogged bool
}

type loginContent struct {
	app.Compo
}

func (p *LoginPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("login / littr")
}

func (p *LoginPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&loginContent{},
		&footer{},
	)
}

func (c *loginContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr login"),
		app.P().Body(
			app.A().Href("/register").Text("don't be mid, join us and be lit! (sign up here)"),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border").Body(
			app.Input().Type("text"),
			app.Label().Text("nickname"),
		),
		app.Div().Class("field label border").Body(
			app.Input().Type("password"),
			app.Label().Text("passphrase"),
		),
		app.Button().Class("responsive primary").Text("login"),
	)
}
