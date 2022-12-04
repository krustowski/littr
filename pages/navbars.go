package pages

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type header struct {
	app.Compo
	updateAvailable bool
}

type footer struct {
	app.Compo
}

const headerString = "littr"

func (h *header) OnAppUpdate(ctx app.Context) {
	// Reports that an app update is available.
	h.updateAvailable = ctx.AppUpdateAvailable()
}

func (h *header) onUpdateClick(ctx app.Context, e app.Event) {
	// Reloads the page to display the modifications.
	ctx.Reload()
}

// top navbar
func (h *header) Render() app.UI {
	return app.Nav().ID("nav").Class("top fixed-top center-align deep-orange5").
		//Style("background-color", navbarColor).
		Body(
			app.A().Href("/settings").Text("settings").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("build")),
				app.Span().Body(
					app.Text("settings")),
			),

			app.H4().Text(headerString).Class("large-padding"),

			// show update button on update
			app.If(h.updateAvailable,
				app.A().Text("update").OnClick(h.onUpdateClick).Body(
					app.I().Class("large").Body(
						app.Text("update")),
					app.Span().Body(
						app.Text("update")),
				),
			),

			app.A().Href("/login").Text("login").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("login")),
				app.Span().Body(
					app.Text("login")),
			),
		)
}

// bottom navbar
func (f *footer) Render() app.UI {
	return app.Nav().ID("nav").Class("bottom fixed-bottom center-align deep-orange5").
		//Style("background-color", navbarColor).
		Body(
			app.A().Href("/stats").Text("stats").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("query_stats")),
				app.Span().Body(
					app.Text("stats")),
			),
			app.A().Href("/users").Text("users").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("group")),
				app.Span().Class("large").Body(
					app.Text("users")),
			),
			app.A().Href("/post").Text("post").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("add")),
				app.Span().Body(
					app.Text("post")),
			),
			app.A().Href("/polls").Text("polls").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("equalizer")),
				app.Span().Body(
					app.Text("polls")),
			),
			app.A().Href("/flow").Text("flow").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("trending_up")),
				app.Span().Body(
					app.Text("flow")),
			),
		)
}
