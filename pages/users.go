package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type UsersPage struct {
	app.Compo
}

func (p *UsersPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		//&usersList{},
		app.Div().Class("large-space"),
		&footer{},
	)
}
