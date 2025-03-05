package navbars

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// onMessage is a callback called when a new message/event is received.
func (h *Header) onMessage(ctx app.Context, e app.Event) {
	// Get the JS data.
	event := common.Event{Data: e.JSValue().Get("data").String(), Type: e.JSValue().Get("eventName").String()}

	// ... => see navbars/action_handlers.go
	ctx.NewActionWithValue("generic-event", event)
}

func (h *Header) onInstallButtonClicked(ctx app.Context, e app.Event) {
	ctx.ShowAppInstallPrompt()
}

func (h *Header) onClickShowLogoutModal(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalLogoutShow = true
	})
}

func (h *Header) onClickModalDismiss(ctx app.Context, e app.Event) {
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
