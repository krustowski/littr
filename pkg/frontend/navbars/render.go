package navbars

import (
	"log"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// top navbar
func (h *Header) Render() app.UI {
	// A very nasty way on how to store the timestamp...
	var last int64 = 0

	// The last beat's timestamp fetch procedure.
	LS := app.Window().Get("localStorage")
	if !LS.IsNull() && !LS.Call("getItem", "lastEventTime").IsNull() {
		str := LS.Call("getItem", "lastEventTime").String()

		lastInt, err := strconv.Atoi(str)
		if err != nil {
			log.Println(err.Error())
		}

		last = int64(lastInt)
	}

	// The very SSE online status (last ~15 seconds).
	sseConnStatus := "disconnected"
	if last > 0 && (time.Now().Unix()-last) < 45 {
		sseConnStatus = "connected"
	}

	// Set the toast default content.
	toastText := h.toast.TText
	if toastText == "" {
		toastText = "new post added to the flow"
	}

	// Link to the settings view.
	settingsHref := "/settings"

	// If not authorized, hide the bar and its items.
	if !h.authGranted {
		settingsHref = "#"
	}

	// Render.
	return app.Nav().ID("nav-top").Class("top fixed-top center-align").Style("opacity", "1.0").
		//Style("background-color", navbarColor).
		Body(
			app.Div().Class("row max shrink").Style("width", "100%").Style("justify-content", "space-between").Body(
				app.If(h.authGranted,
					app.A().Class("button circle transparent").Href(settingsHref).Text("settings").Class("").Title("settings [6]").Aria("label", "settings").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("build")),
					),
				).Else(
					app.Div().Class(""),
				),

				// show intallation button if available
				app.If(h.appInstallable,
					app.A().Class("button circle transparent").Text("install").OnClick(h.onInstallButtonClicked).Title("install").Aria("label", "install").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("download"),
						),
					),
				// hotfix to keep the nav items' distances
				).Else(
					app.Div().Class(""),
				),

				// app logout modal
				app.If(h.modalLogoutShow,
					app.Dialog().ID("logout-modal").Class("grey9 white-text active").Style("border-radius", "8px").Body(
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
							app.Button().Class("max border deep-orange7 white-text").Style("border-radius", "8px").Text("yeah").OnClick(h.onClickLogout),
							app.Button().Class("max border deep-orange7 white-text").Style("border-radius", "8px").Text("nope").OnClick(h.onClickModalDismiss),
						),
					),
				),

				// littr header
				app.Div().Class("row center-align").Body(
					app.H4().Title("system info (click to open)").Class("center-align deep-orange-text").OnClick(h.onClickHeadline).ID("top-header").Body(
						app.Span().Body(
							app.Text(headerString),
							app.If(app.Getenv("APP_ENVIRONMENT") != "prod",
								app.Span().Class("col").Body(
									app.Sup().Body(
										app.Text(" (dev) "),
									),
								),
							),
						),
					),

					// snackbar toast
					app.A().ID("snackbar-general-link").Href("").OnClick(h.onClickModalDismiss).Body(
						app.If(toastText != "",
							app.Div().ID("snackbar-general").Class("snackbar white-text bottom "+common.ToastColor(h.toast.TType)).Body(
							//app.I().Text("info"),
							//app.Span().Text(toastText),
							),
						),
					),
				),

				// app info modal
				app.If(h.modalInfoShow,
					app.Dialog().ID("info-modal").Class("grey9 white-text center-align active").Style("border-radius", "8px").Body(
						app.Article().Class("row center-align").Style("border-radius", "8px").Body(
							app.Img().Src("/web/android-chrome-192x192.png"),
							app.H4().Body(
								app.Span().Body(
									app.Text("littr"),
									app.If(app.Getenv("APP_ENVIRONMENT") != "prod",
										app.Span().Class("col").Body(
											app.Sup().Body(
												app.Text(" (dev) "),
											),
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
								app.A().Class("deep-orange-text bold").Href("https://krusty.space/projects/littr").Text("Documentation (external)"),
							),
						),

						app.Article().Class("center-align").Style("border-radius", "8px").Body(
							app.Text("version: "),
							app.A().Text(app.Getenv("APP_VERSION")).Href("https://github.com/krustowski/littr").Style("font-weight", "bolder"),
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
								app.A().Href("https://github.com/thevxn/swis-api").Text("swapi").Style("font-weight", "bolder"),
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
					app.A().Class("button circle transparent").Text("update").OnClick(h.onClickReload).Title("update").Aria("label", "update").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("update"),
						),
					),
				// hotfix to keep the nav items' distances
				).Else(
					app.A().Class("").OnClick(nil),
				),

				// login/logout button
				app.If(h.authGranted,
					app.A().Class("button circle transparent").Text("logout").Class("").OnClick(h.onClickShowLogoutModal).Title("logout").Aria("label", "logout").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("logout")),
					),
				).Else(
					app.A().Class("button circle transparent").Href("/login").Text("login").Class("").Title("login").Aria("label", "login").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("login")),
					),
				),
			),
		)
}

// bottom navbar
func (f *Footer) Render() app.UI {
	statsHref := "/stats"
	usersHref := "/users"
	postHref := "/post"
	pollsHref := "/polls"
	flowHref := "/flow"

	if !f.authGranted {
		/*statsHref = "#"
		usersHref = "#"
		postHref = "#"
		pollsHref = "#"
		flowHref = "#"*/

		return app.Div()
	}

	//return app.Nav().ID("nav-bottom").Class("bottom fixed-top center-align").Style("opacity", "1.0").
	return app.Nav().ID("nav-bottom").Class("bottom fixed-top").Style("opacity", "1.0").
		Body(
			app.Div().Class("row max shrink").Style("width", "100%").Style("justify-content", "space-between").Body(
				app.A().Class("button circle transparent").Href(statsHref).Text("stats").Class("").Title("stats [1]").Aria("label", "stats").Body(
					app.I().Class("large deep-orange-text").Body(
						app.Text("query_stats")),
				),

				app.A().Class("button circle transparent").Href(usersHref).Text("users").Class("").Title("users [2]").Aria("label", "users").Body(
					app.I().Class("large deep-orange-text").Body(
						app.Text("group")),
				),

				app.A().Class("button circle transparent").Href(postHref).Text("post").Class("").Title("new post/poll [3]").Aria("label", "new post/poll").Body(
					app.I().Class("large deep-orange-text").Body(
						app.Text("add")),
				),

				app.A().Class("button circle transparent").Href(pollsHref).Text("polls").Class("").Title("polls [4]").Aria("label", "polls").Body(
					app.I().Class("large deep-orange-text").Body(
						app.Text("equalizer")),
				),

				app.A().Class("button circle transparent").Href(flowHref).Text("flow").Class("").Title("flow [5]").Aria("label", "flow").Body(
					app.I().Class("large deep-orange-text").Body(
						app.Text("tsunami")),
				),
			),
		)
}
