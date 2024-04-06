package frontend

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ToSPage struct {
	app.Compo
}

type tosContent struct {
	app.Compo

	toastText string
	toastShow bool
}

func (p *ToSPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("ToS / littr")
}

func (p *ToSPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&tosContent{},
	)
}

func (c *tosContent) onClickDismiss(ctx app.Context, e app.Event) {
	c.toastShow = false
	//c.buttonDisabled = false
}

func (c *tosContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("littr ToS (terms of service)"),
				//app.P().Text("let us be serious for a sec nocap"),
			),
		),

		// snackbar
		app.A().OnClick(c.onClickDismiss).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		app.Div().Class("padding responsive").Body(
			app.Ol().Class("extra-line large-text padding").Body(
				app.Li().Text("don't comment on things you got no context to"),
				app.Li().Text("you don't have to comment on every post available"),
				app.Li().Text("don't annoy other fellow flowers"),
				app.Li().Text("don't be rude"),
				app.Li().Text("don't make me tap the sign"),
				app.Li().Text("enjoy the ride"),
			),
		),

		app.Div().Class("large-space"),
		app.Div().Class("label padding responsive").Body(
			app.Article().Class("bottom-align medium transparent padding").Body(
				app.Img().Src("https://i.kym-cdn.com/photos/images/original/001/970/928/ce5.jpg").Class("no-padding absolute center middle").Style("max-width", "90%"),
			),
		),
		app.Div().Class("large-space"),
		app.Div().Class("large-space"),
	)
}
