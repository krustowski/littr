package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type PollsPage struct {
	app.Compo
}

func (p *PollsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		//&pollsList{},
		app.Div().Class("large-space"),
		&footer{},
	)
}
