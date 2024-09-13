package frontend

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	//"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

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

	keyDownEventListener   func()
	eventListenerMessage   func()
	eventListenerKeepAlive func()
	lastHeartbeatTime      int64

	toastText string
	toastShow bool
	toastType string
}

type footer struct {
	app.Compo

	authGranted bool
}

const (
	headerString = "littr"
)

func (h *header) handleDismiss(ctx app.Context, a app.Action) {
	/*deleteModal := app.Window().GetElementByID("delete-modal")
	if !deleteModal.IsNull() {
		deleteModal.Get("classList").Call("remove", "active")
	}

	userModal := app.Window().GetElementByID("user-modal")
	if !userModal.IsNull() {
		userModal.Get("classList").Call("remove", "active")
	}*/

	snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() {
		snack.Get("classList").Call("remove", "active")
	}

	// change title back to the clean one
	title := app.Window().Get("document")
	if !title.IsNull() && strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", prevTitle[4:])
	}

	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = false
		h.modalLogoutShow = false

		h.toastShow = false
		h.toastText = ""
		h.toastType = ""
	})
}

func (h *header) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss-general")
		return
	}

	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	if !authGranted {
		return
	}

	var inputs = []string{
		"post-textarea",
		"poll-question",
		"poll-option-i",
		"poll-option-ii",
		"poll-option-iii",
		"reply-textarea",
		"fig-upload",
		"search",
		"passphrase-current",
		"passphrase-new",
		"passphrase-new-again",
		"about-you-textarea",
		"website-input",
	}

	if helpers.Contains(inputs, app.Window().Get("document").Get("activeElement").Get("id").String()) {
		return
	}

	/*path := ctx.Page().URL().Path

	if path == "/settings" {
	}

	if path == "/flow" {
		replyTextarea := app.Window().GetElementByID("reply-textarea")
		figureInput := app.Window().GetElementByID("fig-upload")

		if !replyTextarea.IsNull() && (len(replyTextarea.Get("value").String()) > 0) {
			return
		}

		if !figureInput.IsNull() && len(figureInput.Get("value").String()) > 0 {
			return
		}
	}

	if path == "/post" {
		postTextarea := app.Window().GetElementByID("post-textarea")

		if !postTextarea.IsNull() && len(postTextarea.Get("value").String()) > 0 {
			return
		}
	}*/

	switch e.Get("key").String() {
	case "1":
		ctx.Navigate("/stats")
	case "2":
		ctx.Navigate("/users")
	case "3":
		ctx.Navigate("/post")
	case "4":
		ctx.Navigate("/polls")
	case "5":
		ctx.Navigate("/flow")
	case "6":
		ctx.Navigate("/settings")
	}
}

func (h *header) onMessage(ctx app.Context, e app.Event) {
	data := e.JSValue().Get("data").String()

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

	// do not parse the message when user has live mode disabled
	/*if !user.LiveMode {
		return
	}*/

	// explode the data CSV string
	slice := strings.Split(data, ",")
	text := ""

	switch slice[0] {
	case "server-stop":
		// server is stoping/restarting
		text = "server is restarting..."
		break

	case "server-start":
		// server is booting up
		text = "server has just started"
		break

	case "post":
		author := slice[1]
		if author == user.Nickname {
			return
		}

		if flowed, found := user.FlowList[author]; !flowed || !found {
			return
		}

		text = "new post added by " + author
		break

	case "poll":
		text = "new poll has been added"
		break
	}

	// show the snack bar the nasty way
	snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() && text != "" {
		snack.Get("classList").Call("add", "active")
		snack.Set("innerHTML", "<i>info</i>"+text)
	}

	// change title to indicate a new event
	title := app.Window().Get("document")
	if !title.IsNull() && !strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", "(*) "+prevTitle)
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
	if !authGranted && path != "/" && path != "/login" && path != "/register" && !strings.Contains(path, "/reset") && path != "/tos" {
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
	h.eventListenerKeepAlive = app.Window().AddEventListener("keepalive", h.onMessage)
	h.keyDownEventListener = app.Window().AddEventListener("keydown", h.onKeyDown)

	ctx.Handle("dismiss-general", h.handleDismiss)

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

func (f *footer) OnMount(ctx app.Context) {
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	f.authGranted = authGranted
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
		if strings.Contains(snack.Get("innerText").String(), "post added") {
			if ctx.Page().URL().Path == "/flow" && !app.Window().GetElementByID("refresh-button").IsNull() {
				ctx.NewAction("dismiss")
				ctx.NewAction("clear")
				ctx.NewAction("refresh")
				return
			}

			ctx.Navigate("/flow")
		}

		if strings.Contains(snack.Get("innerText").String(), "poll added") {
			ctx.Navigate("/polls")
		}

		ctx.NewAction("dismiss-general")
		return
	}

	ctx.NewAction("dismiss-general")
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
		input := callInput{
			Method:      "POST",
			Url:         "/api/v1/auth/logout",
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		if _, ok := littrAPI(input); !ok {
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

	settingsHref := "/settings"

	if !h.authGranted {
		settingsHref = "#"
	}

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
							app.If(app.Getenv("APP_ENVIRONMENT") == "dev",
								app.Span().Class("col").Body(
									app.Sup().Body(
										app.Text(" (dev) "),
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
					app.Dialog().ID("info-modal").Class("grey9 white-text center-align active").Style("border-radius", "8px").Body(
						app.Article().Class("row center-align").Style("border-radius", "8px").Body(
							app.Img().Src("/web/android-chrome-192x192.png"),
							app.H4().Body(
								app.Span().Body(
									app.Text("littr"),
									app.If(app.Getenv("APP_ENVIRONMENT") == "dev",
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
func (f *footer) Render() app.UI {
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
