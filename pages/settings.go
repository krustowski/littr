package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type SettingsPage struct {
	app.Compo
}

type settingsContent struct {
	app.Compo
}

func (p *SettingsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&settingsContent{},
		app.Div().Class("large-space"),
		&footer{},
	)
}

func (c *settingsContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings"),
		app.P().Text("change your passphrase, the about string or just fuck off-"),
		app.Div().Class("space"),
	)
}
