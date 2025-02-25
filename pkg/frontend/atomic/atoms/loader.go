package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Loader struct {
	app.Compo

	ID string

	ShowLoader bool
}

func (l *Loader) Render() app.UI {
	return app.Div().ID(l.ID).Body(
		app.If(l.ShowLoader, func() app.UI {
			return app.Div().Body(
				app.Div().Class("small-space"),
				app.Progress().Class("circle center large deep-orange-border active"),
			)
		}),
	)
}
