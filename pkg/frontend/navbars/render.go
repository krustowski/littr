package navbars

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/frontend/atomic/organisms"
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
	h.sseConnStatus = "disconnected"
	if last > 0 && (time.Now().Unix()-last) < 45 {
		h.sseConnStatus = "connected"
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

	// Top navbar render.
	return app.Nav().ID("nav-top").Class("top fixed-top center-align").Style("opacity", "1.0").
		//Style("background-color", navbarColor).
		Body(
			app.Div().Class("row top-items max shrink").Style("width", "100%").Style("justify-content", "space-between").Body(
				app.If(h.authGranted, func() app.UI {
					return &atoms.Button{
						ID:                "",
						Class:             "circle transparent blue-text",
						Title:             "settings [6]",
						Aria:              map[string]string{"label": "setings"},
						Icon:              "build",
						OnClickActionName: "settings-click",
					}
				}).Else(func() app.UI {
					return app.A().Class("").OnClick(nil).Body()
				}),

				// show intallation button if available
				app.If(h.appInstallable, func() app.UI {
					return &atoms.Button{
						ID:                "",
						Class:             "circle transparent blue-text",
						Title:             "install",
						Aria:              map[string]string{"label": "install"},
						Icon:              "download",
						OnClickActionName: "install-click",
					}
				}).Else(func() app.UI {
					// hotfix to keep the nav items' distances
					return app.A().Class("").OnClick(nil).Body()
				}),

				&organisms.ModalUserLogout{
					User:                     h.user,
					ShowModal:                h.modalLogoutShow,
					OnClickDismissActionName: "dismiss-general",
					OnClickLogoutActionName:  "logout",
					OnClickFlowActionName:    "user-flow-click",
				},

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

				&organisms.ModalAppInfo{
					ShowModal:                h.modalInfoShow,
					SseConnectionStatus:      h.sseConnStatus,
					OnClickDismissActionName: "dismiss-general",
					OnClickReloadActionName:  "reload",
				},

				// Update button
				app.If(h.updateAvailable, func() app.UI {
					return &atoms.Button{
						BadgeText:         "NEW",
						ID:                "",
						Class:             "circle transparent blue-text",
						Title:             "update available",
						Aria:              map[string]string{"label": "update"},
						Icon:              "update",
						OnClickActionName: "reload",
					}
				}).Else(func() app.UI {
					// hotfix to keep the nav items' distances
					return app.A().Class("").OnClick(nil).Body()
				}),

				// login/logout button
				app.If(h.authGranted, func() app.UI {
					return &atoms.Button{
						ID:                "",
						Class:             "circle transparent blue-text",
						Title:             "user info",
						Aria:              map[string]string{"label": "user_info"},
						Icon:              "person",
						OnClickActionName: "user-modal-show",
					}
				}).Else(func() app.UI {
					return &atoms.Button{
						ID:                "",
						Class:             "circle transparent blue-text",
						Title:             "login link",
						Aria:              map[string]string{"label": "login"},
						Icon:              "key_vertical",
						OnClickActionName: "login-click",
					}
				}),
			),
		)
}

// bottom navbar
func (f *Footer) Render() app.UI {
	if !f.authGranted {
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
	}

	//return app.Nav().ID("nav-bottom").Class("bottom fixed-top center-align").Style("opacity", "1.0").
	return app.Nav().ID("nav-bottom").Class("bottom fixed-top").Style("opacity", "1.0").
		Body(
			app.Div().Class("row max shrink").Style("width", "100%").Style("justify-content", "space-between").Body(
				&atoms.Button{
					ID:                "button-stats",
					Class:             "circle transparent blue-text",
					Title:             "stats [1]",
					Aria:              map[string]string{"label": "stats"},
					Icon:              "query_stats",
					OnClickActionName: "stats-click",
				},

				&atoms.Button{
					BadgeText:         fmt.Sprintf("%d", reqCount()),
					ID:                "button-users",
					Class:             "circle transparent blue-text",
					Title:             "users [2]",
					Aria:              map[string]string{"label": "users"},
					Icon:              "group",
					OnClickActionName: "users-click",
				},

				&atoms.Button{
					ID:                "button-post",
					Class:             "circle transparent blue-text",
					Title:             "post [3]",
					Aria:              map[string]string{"label": "post"},
					Icon:              "add",
					OnClickActionName: "post-click",
				},

				&atoms.Button{
					ID:                "button-polls",
					Class:             "circle transparent blue-text",
					Title:             "polls [4]",
					Aria:              map[string]string{"label": "polls"},
					Icon:              "equalizer",
					OnClickActionName: "polls-click",
				},

				&atoms.Button{
					ID:                "button-flow",
					Class:             "circle transparent blue-text",
					Title:             "flow [5]",
					Aria:              map[string]string{"label": "flow"},
					Icon:              "tsunami",
					OnClickActionName: "flow-click",
				},
			),
		)
}
