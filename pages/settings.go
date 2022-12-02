package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type SettingsPage struct {
	app.Compo
}

func (p *SettingsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		//&settingsTable{},
		app.Div().Class("large-space"),
		&footer{},
	)
}
