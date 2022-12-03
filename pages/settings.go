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
		&footer{},
	)
}

func (c *settingsContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings"),
		app.P().Text("change your passphrase, the about string or just fuck off-"),
		app.Div().Class("space"),

		app.Div().Class("field label border").Body(
			app.Input().Type("password"),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border").Body(
			app.Input().Type("password"),
			app.Label().Text("passphrase again"),
		),
		app.Button().Class("responsive primary").Text("change passphrase"),
		app.Div().Class("large-divider"),

		app.Div().Class("field textarea label border extra").Body(
			app.Textarea().Text("idk"),
			app.Label().Text("about"),
		),
		app.Button().Class("responsive primary").Text("change about"),

		app.Div().Class("large-space"),
		app.Div().Class("large-space"),
	)
}
