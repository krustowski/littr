package navbars

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (h *Header) onKeyDown(ctx app.Context, e app.Event) {
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

func (h *Header) onMessage(ctx app.Context, e app.Event) {
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

	// won't trigger the render for some reason... see the bypass ^^
	/*ctx.Dispatch(func(ctx app.Context) {
		//h.toastText = "new post added above"
		//h.toastType = "info"
	})*/
	return
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
			toast.Text("cannot reach backend").Type("error").Dispatch(h, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(h, dispatch)
			return
		}

		ctx.Navigate("/logout")
	})
}
