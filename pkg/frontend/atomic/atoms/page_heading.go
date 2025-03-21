package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type PageHeading struct {
	app.Compo

	Level int
	Title string
	Class string
}

func (p *PageHeading) composeClass() string {
	if p.Class != "" {
		return p.Class
	}

	return "max padding"
}

func (p *PageHeading) composeHeading() app.UI {
	switch p.Level {
	case 1:
		return app.H1().Text(p.Title)
	case 2:
		return app.H2().Text(p.Title)
	case 3:
		return app.H3().Text(p.Title)
	case 4:
		return app.H4().Text(p.Title)
	case 6:
		return app.H6().Text(p.Title)
	default:
		return app.H5().Text(p.Title)
	}
}

func (p *PageHeading) Render() app.UI {
	return app.Div().Class("row").Body(
		app.Div().Class(p.composeClass()).Body(
			p.composeHeading(),
		),
		app.Div().Class("space"),
	)
}
