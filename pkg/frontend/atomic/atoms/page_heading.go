package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type PageHeading struct {
	app.Compo

	Title string
	Class string
}

func (p *PageHeading) composeClass() string {
	if p.Class != "" {
		return p.Class
	}

	return "max padding"
}

func (p *PageHeading) Render() app.UI {
	return app.Div().Class("row").Body(
		app.Div().Class(p.composeClass()).Body(
			app.H5().Text(p.Title),
		),
		app.Div().Class("space"),
	)
}
