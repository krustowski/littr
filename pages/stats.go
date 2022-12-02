package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type StatsPage struct {
	app.Compo
}

func (p *StatsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		//&statsTable{},
		app.Div().Class("large-space"),
		&footer{},
	)
}
