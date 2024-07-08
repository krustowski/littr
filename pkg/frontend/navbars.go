package frontend

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"go.savla.dev/littr/configs"
	"go.savla.dev/littr/pkg/models"

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
	lastHeartbeatTime    int64

	toastText string
	toastShow bool
	toastType string
}

type footer struct {
	app.Compo
}

const (
	headerString = "littr"
)

func (h *header) onMessage(ctx app.Context, e app.Event) {
	data := e.JSValue().Get("data").String()

	//ctx.LocalStorage().Set("lastEventTime", time.Now().UnixNano())

	if data == "heartbeat" {
		return
	}

	var baseString string
	var user models.User
	ctx.LocalStorage().Get("user", &baseString)

	str, err := base64.StdEncoding.DecodeString(baseString)
	if err != nil {
		log.Println(err.Error())
	}

	err = json.Unmarshal(str, &user)
	if err != nil {
		log.Println(err.Error())
	}

	if data == user.Nickname {
		return
	}

	if _, flowed := user.FlowList[data]; !flowed {
		return
	}

	// show the snack bar the nasty way
	snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() {
		snack.Get("classList").Call("add", "active")
	}

	/*ctx.Dispatch(func(ctx app.Context) {
		// won't trigger the render for some reason...
		//h.toastText = "new post added above"
		//h.toastType = "info"
	})*/
	return
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
	path := ctx.Page().URL().Path
	if !authGranted && path != "/login" && path != "/register" && path != "/reset" && path != "/tos" {
		ctx.Navigate("/login")
		return
	}

	// redirect auth'd client from the unauthorized zone
	if authGranted && (path == "/" || path == "/login" || path == "/register" || path == "/reset") {
		ctx.Navigate("/flow")
		return
	}

	// create event listener for SSE messages
	h.eventListenerMessage = app.Window().AddEventListener("message", h.onMessage)

	ctx.Dispatch(func(ctx app.Context) {
		h.authGranted = authGranted
		h.pagePath = path
		//h.toastText = "lmaooooo"
	})

	// keep the update button on until clicked
	var newUpdate bool
	ctx.LocalStorage().Get("newUpdate", &newUpdate)

	if newUpdate {
		h.updateAvailable = true
	}

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
	snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() {
		snack.Get("classList").Call("remove", "active")
	}

	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = false
		h.modalLogoutShow = false

		h.toastShow = false
		h.toastText = ""
		h.toastType = ""
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

	ctx.Async(func() {
		if _, ok := litterAPI("POST", "/api/v1/auth/logout", nil, "", 0); !ok {
			toastText := "cannot POST logout request"

			ctx.Dispatch(func(ctx app.Context) {
				h.toastText = toastText
				h.toastShow = (toastText != "")
			})
			return
		}
	})

	ctx.Navigate("/logout")
}

// top navbar
func (h *header) Render() app.UI {
	toastColor := ""

	switch h.toastType {
	case "success":
		toastColor = "green10"
		break

	case "error":
		toastColor = "red10"
		break

	default:
		toastColor = "blue10"
	}

	// a very nasty way on how to store the timestamp...
	var last int64 = 0

	beat := app.Window().Get("localStorage")
	if !beat.IsNull() && !beat.Call("getItem", "lastEventTime").IsNull() {
		str := beat.Call("getItem", "lastEventTime").String()

		lastInt, err := strconv.Atoi(str)
		if err != nil {
			log.Println(err.Error())
		}

		last = int64(lastInt)
	}

	sseConnStatus := "disconnected"
	if last > 0 && (time.Now().Unix()-last) < 45 {
		sseConnStatus = "connected"
	}

	toastText := h.toastText
	if toastText == "" {
		toastText = "new post added to the flow"
	}

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
				app.H4().Class("center-align deep-orange-text").OnClick(h.onClickHeadline).ID("top-header").Body(
					app.Span().Body(
						app.Text(headerString),
						app.If(configs.APP_ENVIRONMENT == "dev",
						app.Span().Class("col").Body(
							app.Sup().Body(
								app.Text(" (beta) "),
							),
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

				// snackbar toast
				//app.If(h.toastText != "",
				app.Div().ID("snackbar-general").OnClick(h.onClickModalDismiss).Class("snackbar white-text top "+toastColor).Body(
					app.I().Text("error"),
					app.Span().Text(toastText),
				),
				//),
			),

			// app info modal
			app.If(h.modalInfoShow,
				app.Dialog().Class("grey9 white-text center-align active").Style("border-radius", "8px").Body(
					app.Article().Class("row center-align").Style("border-radius", "8px").Body(
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
					app.Article().Class("center-align large-text").Style("border-radius", "8px").Body(
						app.P().Body(
							app.A().Class("deep-orange-text bold").Href("/tos").Text("Terms of Service"),
						),
						app.P().Body(
							app.A().Class("deep-orange-text bold").Href("https://krusty.space/projects/litter").Text("Lore and overview article"),
						),
					),

					app.Article().Class("center-align").Style("border-radius", "8px").Body(
						app.Text("version: "),
						app.A().Text("v"+app.Getenv("APP_VERSION")).Href("https://github.com/krustowski/litter-go").Style("font-weight", "bolder"),
						app.P().Body(
							app.Text("SSE status: "),
							app.If(sseConnStatus == "connected",
								app.Span().ID("heartbeat-info-text").Text(sseConnStatus).Class("green-text bold"),
							).Else(
								app.Span().ID("heartbeat-info-text").Text(sseConnStatus).Class("amber-text bold"),
							),
						),
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
