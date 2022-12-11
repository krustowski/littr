package pages

import (
	"litter-go/backend"
	"os"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type header struct {
	app.Compo

	updateAvailable bool
	appInstallable  bool

	userLogged bool
	userStruct backend.User

	modalShow bool
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

func (h *header) OnMount(ctx app.Context) {
	h.appInstallable = ctx.IsAppInstallable()

	ctx.LocalStorage().Get("userLogged", &h.userLogged)

	if !h.userLogged && ctx.Page().URL().Path != "/login" && ctx.Page().URL().Path != "/register" {
		ctx.Navigate("/login")
	}
}

func (h *header) OnAppInstallChange(ctx app.Context) {
	h.appInstallable = ctx.IsAppInstallable()
}

func (h *header) onInstallButtonClicked(ctx app.Context, e app.Event) {
	ctx.ShowAppInstallPrompt()
}

func (h *header) onClickHeadline(ctx app.Context, e app.Event) {
	h.modalShow = true
}
func (h *header) onClickModalDismiss(ctx app.Context, e app.Event) {
	h.modalShow = false
}

// top navbar
func (h *header) Render() app.UI {
	modalActiveClass := ""
	if h.modalShow {
		modalActiveClass = " active"
	}

	return app.Nav().ID("nav").Class("top fixed-top center-align deep-orange5").
		//Style("background-color", navbarColor).
		Body(
			app.A().Href("/settings").Text("settings").Class("max").Body(
				app.I().Class("large").Body(
					app.Text("build")),
				app.Span().Body(
					app.Text("settings")),
			),

			app.If(h.appInstallable,
				app.A().Text("install").OnClick(h.onInstallButtonClicked).Body(
					app.I().Class("large").Body(
						app.Text("download"),
					),
					app.Span().Body(
						app.Text("install"),
					),
				),
			),

			app.H4().Text(headerString).Class("large-padding").OnClick(h.onClickHeadline),

			// app modal
			app.Div().Class("modal deep-orange white-text"+modalActiveClass).Body(
				app.H5().Text("litter-go (littr) PWA"),
				app.P().Text("version v"+string(os.Getenv("APP_VERSION"))),
				app.Nav().Class("center-align").Body(
					app.Button().Class("border deep-orange7 white-text").Text("Close").OnClick(h.onClickModalDismiss),
				),
			),

			// show update button on update
			app.If(h.updateAvailable,
				app.A().Text("update").OnClick(h.onUpdateClick).Body(
					app.I().Class("large").Body(
						app.Text("update"),
					),
					app.Span().Body(
						app.Text("update"),
					),
				),
			),

			app.If(h.userLogged,
				app.A().Href("/logout").Text("logout").Class("max").Body(
					app.I().Class("large").Body(
						app.Text("logout")),
					app.Span().Body(
						app.Text("logout")),
				),
			).Else(
				app.A().Href("/login").Text("login").Class("max").Body(
					app.I().Class("large").Body(
						app.Text("login")),
					app.Span().Body(
						app.Text("login")),
				),
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
