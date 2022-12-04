package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type PollsPage struct {
	app.Compo
}

type pollsContent struct {
	app.Compo
}

func (p *PollsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("polls / littr")
}

func (p *PollsPage) Render() app.UI {
	return app.Div().Body(
		//app.Body().Class("dark"),
		&header{},
		&pollsContent{},
		&footer{},
	)
}

func (c *pollsContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr polls"),
		app.P().Text("to be implemented soon"),
	)
}
