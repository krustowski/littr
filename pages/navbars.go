package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

// calm pale green
//const navbarColor = "#206040"
const navbarColor = "#cc6600"

type header struct {
	app.Compo
}

type footer struct {
	app.Compo
}

// top navbar
func (h *header) Render() app.UI {
	return app.Nav().ID("nav").Class("top fixed-top center-align").Style("background-color", navbarColor).Body(
		app.A().Href("/settings").Text("settings").Body(
			app.I().Body(
				app.Text("build")),
			app.Span().Body(
				app.Text("settings")),
		),
		app.H4().Text("littr").Class("large-padding max"),
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
	return app.Nav().ID("nav").Class("bottom fixed-bottom center-align").Style("background-color", navbarColor).Body(
		app.A().Href("/stats").Text("stats").Body(
			app.I().Body(
				app.Text("query_stats")),
			app.Span().Body(
				app.Text("stats")),
		),
		app.A().Href("/users").Text("users").Class("max").Body(
			app.I().Body(
				app.Text("group")),
			app.Span().Body(
				app.Text("users")),
		),
		app.A().Href("/post").Text("post").Body(
			app.I().Body(
				app.Text("add")),
			app.Span().Body(
				app.Text("post")),
		),
		app.A().Href("/polls").Text("polls").Class("max").Body(
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
