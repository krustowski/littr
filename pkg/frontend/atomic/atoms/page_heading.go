package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type PageHeading struct {
	app.Compo

	Title string
}

func (p *PageHeading) Render() app.UI {
	return app.Div().Class("row").Body(
		app.Div().Class("max padding").Body(
			app.H5().Text(p.Title),
		),
		app.Div().Class("space"),
	)
}
