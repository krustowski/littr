package navbars

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

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
	toastTopText := h.toastTop.TText
	if toastTopText == "" {
		toastTopText = "new action just dropped"
	}

	toastBottomText := h.toastBottom.TText
	if toastBottomText == "" {
		toastBottomText = "new post added to the flow"
	}

	// Link to the settings view.
	settingsHref := "/settings"

	// If not authorized, hide the bar and its items.
	if !h.authGranted {
		settingsHref = "#"
	}

	// Top navbar render.
	return app.Nav().ID("nav-top").Class("top fixed-top center-align").Style("opacity", "1.0").
		//Style("background-color", navbarColor).
		Body(
			app.Div().Class("row max shrink").Style("width", "100%").Style("justify-content", "space-between").Body(
				app.If(h.authGranted, func() app.UI {
					return app.A().Class("button circle transparent").Href(settingsHref).Text("settings").Class("").Title("settings [6]").Aria("label", "settings").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("build")),
					)
				}).Else(func() app.UI {
					return app.A().Class("").OnClick(nil).Body()
				}),

				// show intallation button if available
				app.If(h.appInstallable, func() app.UI {
					return app.A().Class("button circle transparent").Text("install").OnClick(h.onInstallButtonClicked).Title("install").Aria("label", "install").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("download"),
						),
					)
					// hotfix to keep the nav items' distances
				}).Else(func() app.UI {
					return app.A().Class("").OnClick(nil).Body()
				}),

				// app logout modal
				app.If(h.modalLogoutShow, func() app.UI {
					return app.Dialog().ID("logout-modal").Class("grey10 white-text active thicc").Body(
						app.Nav().Class("center-align").Body(
							app.H5().Text("user"),
						),

						app.Div().Class("space"),

						// User's avatar and nickname.
						app.Div().Class("row border thicc").Body(
							app.Img().ID(h.user.Nickname).Title("user's avatar").Class("responsive padding max left").Src(h.user.AvatarURL).Style("max-width", "100px").Style("border-radius", "50%").OnClick(h.onClickUserFlow),
							app.P().Class("max").Body(
								app.A().Title("user's flow link").Class("bold deep-orange-text").OnClick(h.onClickUserFlow).ID(h.user.Nickname).Body(
									app.Span().ID("user-flow-link").Class("large-text bold deep-orange-text").Text(h.user.Nickname),
								),
							),
							app.Button().ID(h.user.Nickname).Class("max bold deep-orange7 white-text thicc").OnClick(h.onClickUserFlow).Style("margin-right", "15px").Body(
								app.I().Text("tsunami"),
								app.Text("Flow"),
							),
						),

						app.Div().Class("space"),

						/*app.Article().Class("row warn amber-border white-text border thicc").Body(
							app.I().Text("warning").Class("amber-text"),
							app.P().Class("max bold").Body(
								app.Span().Text("Are you sure you want to end this session and log out?"),
							),
						),

						app.Div().Class("space"),*/

						app.Div().Class("row").Body(
							app.Button().Class("max bold black white-text thicc").OnClick(h.onClickModalDismiss).Body(
								app.Span().Body(
									app.I().Style("padding-right", "5px").Text("close"),
									app.Text("Close"),
								),
							),
							app.Button().Class("max bold deep-orange7 white-text thicc").OnClick(h.onClickLogout).Body(
								app.Span().Body(
									app.I().Style("padding-right", "5px").Text("logout"),
									app.Text("Log out"),
								),
							),
						),
					)
				}),

				// littr header
				app.Div().Class("row center-align").Body(
					&atoms.Snackbar{
						Class:    "snackbar white-text thicc",
						ID:       "snackbar-general-top",
						IDLink:   "snackbar-general-top-link",
						Position: "top",
						Text:     toastTopText,
					},

					&molecules.LittrHeader{
						HeaderString:              headerString,
						OnClickHeadlineActionName: "littr-header-click",
					},

					&atoms.Snackbar{
						Class:    "snackbar white-text thicc",
						ID:       "snackbar-general-bottom",
						IDLink:   "snackbar-general-bottom-link",
						Position: "bottom",
						Text:     toastBottomText,
					},
				),

				// app info modal
				app.If(h.modalInfoShow, func() app.UI {
					return app.Dialog().ID("info-modal").Class("grey10 white-text center-align active thicc").Body(
						app.Article().Class("row white-text center-align border thicc").Body(
							app.Img().Src("/web/android-chrome-512x512.svg").Style("max-width", "10em"),
							app.H4().Body(
								app.Span().Body(
									app.Text("littr"),
									app.If(app.Getenv("APP_ENVIRONMENT") != "prod", func() app.UI {
										return app.Span().Class("col").Body(
											app.Sup().Body(
												app.If(app.Getenv("APP_ENVIRONMENT") == "stage", func() app.UI {
													return app.Text(" (stage) ")
												}).Else(func() app.UI {
													return app.Text(" (dev) ")
												}),
											),
										)
									}),
								),
							),
						),

						app.Article().Class("center-align large-text border thicc").Body(
							app.P().Body(
								app.A().Class("deep-orange-text bold").Href("/tos").Text("Terms of Service"),
							),
							app.P().Body(
								app.A().Class("deep-orange-text bold").Href("https://krusty.space/projects/littr").Text("Documentation (external)"),
							),
						),

						app.Article().Class("center-align white-text border thicc").Body(
							app.Text("Version: "),
							app.A().Text(app.Getenv("APP_VERSION")).Href("https://github.com/krustowski/littr").Style("font-weight", "bolder"),
							app.P().Body(
								app.Text("SSE status: "),
								app.If(sseConnStatus == "connected", func() app.UI {
									return app.Span().ID("heartbeat-info-text").Text(sseConnStatus).Class("green-text bold")
								}).Else(func() app.UI {
									return app.Span().ID("heartbeat-info-text").Text(sseConnStatus).Class("amber-text bold")
								}),
							),
						),

						app.Nav().Class("center-align").Body(
							app.P().Body(
								app.Text("Powered by "),
								app.A().Href("https://go-app.dev/").Text("go-app").Style("font-weight", "bolder"),
								app.Text(" & "),
								app.A().Href("https://www.beercss.com/").Text("beercss").Style("font-weight", "bolder"),
							),
						),

						app.Div().Class("row").Body(
							app.Button().Class("max bold black white-text thicc").OnClick(h.onClickModalDismiss).Body(
								app.Span().Body(
									app.I().Style("padding-right", "5px").Text("close"),
									app.Text("Close"),
								),
							),
							app.Button().Class("max bold deep-orange7 white-text thicc").OnClick(h.onClickReload).Body(
								app.Span().Body(
									app.I().Style("padding-right", "5px").Text("refresh"),
									app.Text("Reload"),
								),
							),
						),
					)
				}),

				// update button
				app.If(h.updateAvailable, func() app.UI {
					return app.A().Class("button circle transparent").Text("update").OnClick(h.onClickReload).Title("update").Aria("label", "update").Body(
						app.Div().Class("badge blue-border blue-text border").Text("NEW"),
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("update"),
						),
					)
					// hotfix to keep the nav items' distances
				}).Else(func() app.UI {
					return app.A().Class("").OnClick(nil).Body()
				}),

				// login/logout button
				app.If(h.authGranted, func() app.UI {
					return app.A().Class("button circle transparent").Text("user").Class("").OnClick(h.onClickShowLogoutModal).Title("user").Aria("label", "user").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("person")),
					)
				}).Else(func() app.UI {
					return app.A().Class("button circle transparent").Href("/login").Text("login").Class("").Title("login").Aria("label", "login").Body(
						app.I().Class("large").Class("deep-orange-text").Body(
							app.Text("key_vertical")),
					)
				}),
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

	var reqCount = func() int64 {
		var count int64

		if reflect.DeepEqual(f.user, (models.User{})) || f.user.RequestList == nil {
			return count
		}

		for _, val := range f.user.RequestList {
			if val {
				count++
			}
		}
		return count
	}()

	//return app.Nav().ID("nav-bottom").Class("bottom fixed-top center-align").Style("opacity", "1.0").
	return app.Nav().ID("nav-bottom").Class("bottom fixed-top").Style("opacity", "1.0").
		Body(
			app.Div().Class("row max shrink").Style("width", "100%").Style("justify-content", "space-between").Body(
				app.A().Class("button circle transparent").Href(statsHref).Text("stats").Class("").Title("stats [1]").Aria("label", "stats").Body(
					app.I().Class("large deep-orange-text").Body(
						app.Text("query_stats")),
				),

				app.A().Class("button circle transparent").Href(usersHref).Text("users").Class("").Title("users [2]").Aria("label", "users").Body(
					app.If(reqCount > 0, func() app.UI {
						return app.Div().Class("badge border").Text(fmt.Sprintf("%d", reqCount))
					}),
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
