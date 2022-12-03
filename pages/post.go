package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type PostPage struct {
	app.Compo
}

type postContent struct {
	app.Compo
}

func (p *PostPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("post / littr")
}

func (p *PostPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&postContent{},
		&footer{},
	)
}

func (c *postContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("new flow post"),
		app.P().Text("pleasure to be lit"),

		app.Div().Class("field textarea label border extra").Body(
			app.Textarea(),
			app.Label().Text("contents"),
		),
		app.Button().Class("responsive primary").Text("post"),
	)
}
