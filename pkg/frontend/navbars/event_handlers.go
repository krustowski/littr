package navbars

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/helpers"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// onKeyDown is a callback to handle the key-down event: this allows one to control the app using their keyboard more effectively.
func (h *Header) onKeyDown(ctx app.Context, e app.Event) {
	// Was the key Escape/Esc? Then cast general item dismissal.
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss-general")
		return
	}

	// Fetch the auth state.
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	// Do not continue when unacthenticated/unauthorized.
	if !authGranted {
		return
	}

	// List of inputs, that blocks the refresh event.
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

	// If any input is active (is written in for example), then do not register the R key.
	if helpers.Contains(inputs, app.Window().Get("document").Get("activeElement").Get("id").String()) {
		return
	}

	// Use keys 1-6 to navigate through the UI.
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

// onMessage is a callback called when a new message/event is received.
func (h *Header) onMessage(ctx app.Context, e app.Event) {
	// Get the JS data.
	_ = e.JSValue().Get("data").String()

	// ... => see navbars/action_handlers.go
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

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(h, dispatch)
			return
		}

		ctx.Navigate("/logout")
	})
}
