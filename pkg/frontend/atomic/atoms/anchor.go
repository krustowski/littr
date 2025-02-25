package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Anchor struct {
	app.Compo

	ShowLoader bool
}

func (a *Anchor) Render() app.UI {
	return app.Div().ID("page-end-anchor").Body(
		app.If(a.ShowLoader, func() app.UI {
			return app.Div().Body(
				app.Div().Class("small-space"),
				app.Progress().Class("circle center large deep-orange-border active"),
			)
		}),
	)
}
