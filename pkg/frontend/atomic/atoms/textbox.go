package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type TextBox struct {
	app.Compo

	Class     string
	Icon      string
	IconClass string
	Text      string
}

func (t *TextBox) Render() app.UI {
	return app.Article().Class(t.Class).Body(
		app.I().Text(t.Icon).Class(t.IconClass),
		app.P().Class("max bold").Body(
			app.Span().Text(t.Text),
		),
	)
}
