package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type LoginPage struct {
	app.Compo
}

func (p *LoginPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		//&loginTable{},
		app.Div().Class("large-space"),
		&footer{},
	)
}
