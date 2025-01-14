package navbars

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
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

func (h *Header) onClickHeadline(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = true
	})
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

func (h *Header) onClickReload(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.updateAvailable = false
	})

	ctx.LocalStorage().Set("newUpdate", false)

	ctx.Reload()
}

func (h *Header) onClickLogout(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		h.authGranted = false
	})

	ctx.LocalStorage().Set("user", "")
	ctx.LocalStorage().Set("authGranted", false)

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		input := &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/auth/logout",
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(h, dispatch)
			return
		}

		/*if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(h, dispatch)
			return
		}*/

		ctx.Navigate("/logout")
	})
}

func (h *Header) onClickUserFlow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	if key == "" {
		return
	}

	if strings.Contains(ctx.Page().URL().Path, key) {
		ctx.NewAction("dismiss-general")
		return
	}

	/*if !strings.Contains(ctx.Page().URL().Path, key) && strings.Contains(ctx.Page().URL().Path, "flow") {
		//ctx.NewAction("dismiss-general")
		h.modalLogoutShow = false
	}*/

	/*ctx.Dispatch(func(ctx app.Context) {
		h.modalLogoutShow = false
	})*/

	ctx.Navigate("/flow/users/" + key)

	ctx.Defer(func(app.Context) {
		h.modalLogoutShow = false
	})
}
