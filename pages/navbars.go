package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

const navbarColor = "#206040"

type header struct {
	app.Compo
}

type footer struct {
	app.Compo
}

// top navbar
func (h *header) Render() app.UI {
	return app.Nav().ID("nav").Class("top fixed-top").Style("background-color", navbarColor).Body(
		app.A().Href("/stats").Text("stats").Body(
			app.I().Body(
				app.Text("query_stats")),
			app.Span().Body(
				app.Text("stats")),
		),
		app.H5().Text("littr"),
		app.A().Href("/login").Text("login").Body(
			app.I().Body(
				app.Text("login")),
			app.Span().Body(
				app.Text("login")),
		),
	)
}

// bottom navbar
func (f *footer) Render() app.UI {
	return app.Nav().ID("nav").Class("bottom fixed-bottom").Style("background-color", navbarColor).Body(
		app.A().Href("/settings").Text("settings").Body(
			app.I().Body(
				app.Text("build")),
			app.Span().Body(
				app.Text("settings")),
		),
		app.A().Href("/users").Text("users").Body(
			app.I().Body(
				app.Text("group")),
			app.Span().Body(
				app.Text("users")),
		),
		app.A().Href("/polls").Text("polls").Body(
			app.I().Body(
				app.Text("equalizer")),
			app.Span().Body(
				app.Text("polls")),
		),
		app.A().Href("/flow").Text("flow").Body(
			app.I().Body(
				app.Text("trending_up")),
			app.Span().Body(
				app.Text("flow")),
		),
	)
}
