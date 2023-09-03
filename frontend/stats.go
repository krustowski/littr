package frontend

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type StatsPage struct {
	app.Compo
}

type statsContent struct {
	app.Compo
	stats []string
}

func (p *StatsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("stats / littr")
}

func (p *StatsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&statsContent{},
	)
}

func (c *statsContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr stats"),
		app.P().Text("wanna know your flow stats? how many you got in the flow and vice versa? yo"),
		app.Div().Class("space"),
	)
}
