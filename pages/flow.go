package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type FlowPage struct {
	app.Compo
}

func (p *FlowPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		//&flowContent{},
		app.Div().Class("large-space"),
		&footer{},
	)
}
