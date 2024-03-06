package frontend

import (
	"encoding/json"

	"go.savla.dev/littr/config"
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

func (h *header) tryCookies(ctx app.Context) bool {
	resp := struct {
		Users map[string]models.User `json:"users" binding:"required"`
	}{}

	if data, ok := litterAPI("POST", "/api/token", nil, "ghost", 0); ok {
		if err := json.Unmarshal(*data, &resp); err != nil {
			return false
		}

		if resp.Users == nil {
			return false
		}

		var user models.User
		for _, user = range resp.Users {
		}

		ctx.SetState("user", user)
		return true
	}

	return false
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

	if authGranted {
		ctx.SetState("user", h.user)
	}

	ctx.SetState("authGranted", authGranted)
	h.authGranted = authGranted

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
	//ctx.LocalStorage().Set("authGranted", false)
	ctx.LocalStorage().Set("user", "")
	h.authGranted = false

	ctx.LocalStorage().Set("authGranted", false)
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
	//return app.Nav().ID("nav-top").Class("top fixed-top center-align deep-orange").
	return app.Nav().ID("nav-top").Class("top fixed-top center-align").Style("opacity", "1.0").
		//Style("background-color", navbarColor).
		Body(
			app.A().Href("/settings").Text("settings").Class("max").Body(
				app.I().Class("large").Class("deep-orange-text").Body(
					app.Text("build")),
				//app.Span().Body(
				//app.Text("settings")),
			),

			// show intallation button if available
			app.If(h.appInstallable,
				app.A().Text("install").OnClick(h.onInstallButtonClicked).Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("download"),
					),
					//app.Span().Body(
					//app.Text("install"),
					//),
				),
			),

			app.Div().Class("row").Body(
				app.H4().Class("large-padding deep-orange-text").OnClick(h.onClickHeadline).Body(
					app.Text(headerString),
					app.Span().Class("small-text middle top-align").Text(" (beta)"),
				),
			),

			// snackbar offline mode
			app.A().OnClick(h.onClickModalDismiss).Body(
				app.If(!h.onlineState,
					app.Div().Class("snackbar red5 white-text top active").Body(
						app.Span().Text("internet connection is gone innit..."),
					),
				),
			),

			// app info modal
			app.Dialog().Class("grey9 white-text center-align"+modalInfoActiveClass).Body(
				app.Div().Class("row").Body(
					app.Img().Src("/web/android-chrome-192x192.png"),
					app.H4().Text("littr (beta)"),
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
				app.Div().Class("small-space"),

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

			app.If(h.authGranted,
				app.A().Text("logout").Class("max").OnClick(h.onClickShowLogoutModal).Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("logout")),
					//app.Span().Body(
					//app.Text("logout")),
				),
			).Else(
				app.A().Href("/login").Text("login").Class("max").Body(
					app.I().Class("large").Class("deep-orange-text").Body(
						app.Text("login")),
					//app.Span().Body(
					//app.Text("login")),
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
	//return app.Nav().ID("nav-bottom").Class("bottom fixed-bottom center-align deep-orange8").
	return app.Nav().ID("nav-top").Class("bottom fixed-top center-align").Style("opacity", "1.0").
		Body(
			app.A().Href("/stats").Text("stats").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("query_stats")),
				//app.Span().Body(
				//app.Text("stats")),
			),
			app.A().Href("/users").Text("users").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("group")),
				//app.Span().Class("large").Body(
				//app.Text("users")),
			),
			app.A().Href("/post").Text("post").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("add")),
				//app.Span().Body(
				//app.Text("post")),
			),
			app.A().Href("/polls").Text("polls").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					app.Text("equalizer")),
				//app.Span().Body(
				//app.Text("polls")),
			),
			app.A().Href("/flow").Text("flow").Class("max").Body(
				app.I().Class("large deep-orange-text").Body(
					//app.Text("trending_up")),
					app.Text("tsunami")),
				//app.Span().Body(
				//app.Text("flow")),
			),
		)
}
