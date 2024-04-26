package frontend

import (
	"strings"

	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type header struct {
	app.Compo

	updateAvailable bool
	appInstallable  bool

	authGranted bool
	user        models.User

	modalInfoShow   bool
	modalLogoutShow bool

	onlineState bool

	pagePath string

	eventListenerMessage func()
}

type footer struct {
	app.Compo
}

const (
	headerString = "littr"
)

func (h *header) onMessage(ctx app.Context, e app.Event) {
	data := e.JSValue().Get("data").String()

	if data == "heartbeat" {

		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		//h.toastText = "new post added above"
		//h.toastType = "info"
	})
}

func (h *header) OnAppUpdate(ctx app.Context) {
	// Reports that an app update is available.
	//h.updateAvailable = ctx.AppUpdateAvailable()

	ctx.Dispatch(func(ctx app.Context) {
		h.updateAvailable = true
	})

	ctx.LocalStorage().Set("newUpdate", true)

	// force reload the app on update
	//ctx.Reload()
}

func (h *header) OnMount(ctx app.Context) {
	h.appInstallable = ctx.IsAppInstallable()
	h.onlineState = true

	//authGranted := h.tryCookies(ctx)
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	// redirect client to the unauthorized zone
	if !authGranted && ctx.Page().URL().Path != "/login" && ctx.Page().URL().Path != "/register" && ctx.Page().URL().Path != "/reset" {
		ctx.Navigate("/login")
		return
	}

	// redirect auth'd client from the unauthorized zone
	if authGranted && (ctx.Page().URL().Path == "/" || ctx.Page().URL().Path == "/login" || ctx.Page().URL().Path == "/register") {
		ctx.Navigate("/flow")
		return
	}

	h.authGranted = authGranted

	h.pagePath = ctx.Page().URL().Path

	// keep the update button on until clicked
	var newUpdate bool
	ctx.LocalStorage().Get("newUpdate", &newUpdate)

	if newUpdate {
		h.updateAvailable = true
	}

	// create event listener for SSE messages
	h.eventListenerMessage = app.Window().AddEventListener("message", h.onMessage)

	h.onlineState = true // this is a guess
	// this may not be implemented
	nav := app.Window().Get("navigator")
	if nav.Truthy() {
		onLine := nav.Get("onLine")
		if !onLine.IsUndefined() {
			h.onlineState = onLine.Bool()
		}
	}

	app.Window().Call("addEventListener", "online", app.FuncOf(func(this app.Value, args []app.Value) any {
		h.onlineState = true
		//call(true)
		return nil
	}))

	app.Window().Call("addEventListener", "offline", app.FuncOf(func(this app.Value, args []app.Value) any {
		h.onlineState = false
		//call(false)
		return nil
	}))
}

func (h *header) OnAppInstallChange(ctx app.Context) {
	ctx.Dispatch(func(ctx app.Context) {
		h.appInstallable = ctx.IsAppInstallable()
	})
}

func (h *header) onInstallButtonClicked(ctx app.Context, e app.Event) {
	ctx.ShowAppInstallPrompt()
}

func (h *header) onClickHeadline(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = true
	})
}

func (h *header) onClickShowLogoutModal(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalLogoutShow = true
	})
}

func (h *header) onClickModalDismiss(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = false
		h.modalLogoutShow = false
	})
}

func (h *header) onClickReload(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.updateAvailable = false
	})

	ctx.LocalStorage().Set("newUpdate", false)

	ctx.Reload()
}

func (h *header) onClickLogout(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.authGranted = false
	})

	ctx.LocalStorage().Set("user", "")
	ctx.LocalStorage().Set("authGranted", false)

	ctx.Navigate("/logout")
}

// top navbar
func (h *header) Render() app.UI {
	return app.Nav().ID("nav-top").Class("top fixed-top center-align").Style("opacity", "1.0").
		//Style("background-color", navbarColor).
		Body(
			app.A().Href("/settings").Text("settings").Class("max").Body(
				app.I().Class("large").Class("deep-orange-text").Body(
					app.Text("build")),
			),

			// show intallation button if available
			app.If(h.appInstallable,
				app.A().Class("max").Text("install").OnClick(h.onInstallButtonClicked).Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("download"),
					),
				),
			// hotfix to keep the nav items' distances
			).Else(
				app.A().Class("max").OnClick(nil),
			),

			// app logout modal
			app.If(h.modalLogoutShow,
				app.Dialog().Class("grey9 white-text active").Style("border-radius", "8px").Body(
					app.Nav().Class("center-align").Body(
						app.H5().Text("logout"),
					),

					app.Article().Class("row surface-container-highest").Body(
						app.I().Text("warning").Class("amber-text"),
						app.P().Class("max").Body(
							app.Span().Text("are you sure you want to end this session and log out?"),
						),
					),
					app.Div().Class("space"),

					app.Div().Class("row").Body(
						app.Button().Class("max border red9 white-text").Style("border-radius", "8px").Text("yeah").OnClick(h.onClickLogout),
						app.Button().Class("max border deep-orange7 white-text").Style("border-radius", "8px").Text("nope").OnClick(h.onClickModalDismiss),
					),
				),
			),

			// littr header
			app.Div().Class("max row center-align").Body(
				app.H4().Class("center-align deep-orange-text").OnClick(h.onClickHeadline).Body(
					app.Span().Body(
						app.Text(headerString),
						app.Span().Class("col").Body(
							app.Sup().Body(
								app.Text(" (beta) "),
							),
							app.If(strings.Contains(h.pagePath, "flow"),
								app.Span().Class("dot"),
							),
						),
					),
				),

				// snackbar offline mode
				app.If(!h.onlineState,
					app.Div().OnClick(h.onClickModalDismiss).Class("snackbar red10 white-text top active").Body(
						app.I().Text("warning").Class("amber-text"),
						app.Span().Text("no internet connection"),
					),
				),
			),

			// app info modal
			app.If(h.modalInfoShow,
				app.Dialog().Class("grey9 white-text center-align active").Style("border-radius", "8px").Body(
					app.Div().Class("row center-align").Body(
						app.Img().Src("/web/android-chrome-192x192.png"),
						app.H4().Body(
							app.Span().Body(
								app.Text("littr"),
								app.Span().Class("col").Body(
									app.Sup().Body(
										app.Text(" (beta) "),
									),
								),
							),
						),
					),
					app.Nav().Class("center-align large-text").Body(
						app.P().Body(
							app.A().Class("deep-orange-text bold").Href("/tos").Text("Terms of Service"),
						),
					),
					app.Nav().Class("center-align large-text").Body(
						app.P().Body(
							app.A().Class("deep-orange-text bold").Href("https://krusty.space/projects/litter").Text("Lore and overview article"),
						),
					),

					app.Article().Class("center-align").Style("border-radius", "8px").Body(
						app.Text("version "),
						app.A().Text("v"+app.Getenv("APP_VERSION")).Href("https://github.com/krustowski/litter-go").Style("font-weight", "bolder"),
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

					app.Div().Class("row").Body(
						app.Button().Class("max border deep-orange7 white-text").Style("border-radius", "8px").Text("reload").OnClick(h.onClickReload),
						app.Button().Class("max border deep-orange7 white-text").Style("border-radius", "8px").Text("close").OnClick(h.onClickModalDismiss),
					),
				),
			),

			// update button
			app.If(h.updateAvailable,
				app.A().Class("max").Text("update").OnClick(h.onClickReload).Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("update"),
					),
				),
			// hotfix to keep the nav items' distances
			).Else(
				app.A().Class("max").OnClick(nil),
			),

			// login/logout button
			app.If(h.authGranted,
				app.A().Text("logout").Class("max").OnClick(h.onClickShowLogoutModal).Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("logout")),
				),
			).Else(
				app.A().Href("/login").Text("login").Class("max").Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("login")),
				),
			),
		)
}

// bottom navbar
func (f *footer) Render() app.UI {
	return app.Nav().ID("nav-top").Class("bottom fixed-top center-align").Style("opacity", "1.0").
		Body(
			app.A().Href("/stats").Text("stats").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("query_stats")),
			),

			app.A().Href("/users").Text("users").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("group")),
			),

			app.A().Href("/post").Text("post").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("add")),
			),

			app.A().Href("/polls").Text("polls").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("equalizer")),
			),

			app.A().Href("/flow").Text("flow").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("tsunami")),
			),
		)
}
