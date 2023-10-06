package frontend

import (
	"go.savla.dev/littr/config"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type header struct {
	app.Compo

	updateAvailable bool
	appInstallable  bool

	userLogged bool

	modalInfoShow   bool
	modalLogoutShow bool
}

type footer struct {
	app.Compo
}

const (
	headerString = "littr"
)

func (h *header) OnAppUpdate(ctx app.Context) {
	// Reports that an app update is available.
	h.updateAvailable = ctx.AppUpdateAvailable()

	// force reload the app on update
	ctx.Reload()
}

func (h *header) OnMount(ctx app.Context) {
	h.appInstallable = ctx.IsAppInstallable()

	var encodedUser string
	ctx.LocalStorage().Get("user", &encodedUser)

	h.userLogged = verifyUser(encodedUser)

	if !h.userLogged && ctx.Page().URL().Path != "/login" && ctx.Page().URL().Path != "/register" {
		ctx.Navigate("/login")
		return
	}

	if h.userLogged && (ctx.Page().URL().Path == "/" || ctx.Page().URL().Path == "/login" || ctx.Page().URL().Path == "/register") {
		ctx.Navigate("/flow")
		return
	}
}

func (h *header) OnAppInstallChange(ctx app.Context) {
	h.appInstallable = ctx.IsAppInstallable()
}

func (h *header) onInstallButtonClicked(ctx app.Context, e app.Event) {
	ctx.ShowAppInstallPrompt()
}

func (h *header) onClickHeadline(ctx app.Context, e app.Event) {
	h.modalInfoShow = true
}

func (h *header) onClickShowLogoutModal(ctx app.Context, e app.Event) {
	h.modalLogoutShow = true
}

func (h *header) onClickModalDismiss(ctx app.Context, e app.Event) {
	h.modalInfoShow = false
	h.modalLogoutShow = false
}

func (h *header) onClickReload(ctx app.Context, e app.Event) {
	ctx.Reload()
}

func (h *header) onClickLogout(ctx app.Context, e app.Event) {
	//ctx.LocalStorage().Set("userLogged", false)
	ctx.LocalStorage().Set("user", "")
	h.userLogged = false

	ctx.Navigate("/logout")
}

func (h *header) Render() app.UI {
	modalInfoActiveClass := ""
	if h.modalInfoShow {
		modalInfoActiveClass = " active"
	}

	modalLogoutActiveClass := ""
	if h.modalLogoutShow {
		modalLogoutActiveClass = " active"
	}

	// top navbar
	return app.Nav().ID("nav-top").Class("top fixed-top center-align deep-orange").
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

			// app info modal
			app.Dialog().Class("grey9 white-text center-align"+modalInfoActiveClass).Body(
				app.Img().Src("/web/android-chrome-192x192.png"),
				app.Nav().Class("center-align").Body(
					app.H5().Text("litter-go (littr) PWA"),
				),
				app.Nav().Class("center-align").Body(
					app.P().Body(
						app.Text("version "),
						app.A().Text("v"+config.Version).Href("https://github.com/krustowski/litter-go").Style("font-weight", "bolder"),
					),
					app.Div().Class("small-space"),
				),

				app.Nav().Class("center-align").Body(
					app.P().Body(
						app.Text("powered by "),
						app.A().Href("https://go-app.dev/").Text("go-app").Style("font-weight", "bolder"),
						app.Text(", "),
						app.A().Href("https://www.beercss.com/").Text("beercss").Style("font-weight", "bolder"),
						app.Text(" & "),
						app.A().Href("https://github.com/savla-dev/swis-api").Text("swapi").Style("font-weight", "bolder"),
					),
				),
				app.Nav().Class("center-align").Body(
					app.Button().Class("border deep-orange7 white-text").Text("reload").OnClick(h.onClickReload),
					app.Button().Class("border deep-orange7 white-text").Text("close").OnClick(h.onClickModalDismiss),
				),
			),

			app.If(h.userLogged,
				app.A().Text("logout").Class("max").OnClick(h.onClickShowLogoutModal).Body(
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

			// app logout modal
			app.Dialog().Class("grey9 white-text"+modalLogoutActiveClass).Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("really logout?"),
				),
				app.Div().Class("large-space"),
				app.Nav().Class("center-align").Body(
					app.Button().Class("border deep-orange7 white-text").Text("yes").OnClick(h.onClickLogout),
					app.Button().Class("border deep-orange7 white-text").Text("nah").OnClick(h.onClickModalDismiss),
				),
			),
		)
}

// bottom navbar
func (f *footer) Render() app.UI {
	return app.Nav().ID("nav-bottom").Class("bottom fixed-bottom center-align deep-orange8").
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
