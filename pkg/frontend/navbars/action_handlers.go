package navbars

import (
	//"fmt"
	//"strings"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	//"github.com/tmaxmax/go-sse"
)

func (h *Header) handleDismiss(ctx app.Context, a app.Action) {
	deleteModal := app.Window().GetElementByID("delete-modal")
	if !deleteModal.IsNull() {
		deleteModal.Get("classList").Call("remove", "active")
	}

	userModal := app.Window().GetElementByID("user-modal")
	if !userModal.IsNull() {
		userModal.Get("classList").Call("remove", "active")
	}

	infoModal := app.Window().GetElementByID("info-modal")
	if !infoModal.IsNull() {
		infoModal.Get("classList").Call("remove", "active")
	}

	replyModal := app.Window().GetElementByID("reply-modal")
	if !replyModal.IsNull() {
		replyModal.Get("classList").Call("remove", "active")
	}

	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = false
		h.modalLogoutShow = false

		h.toastShow = false
		h.toastText = ""
		h.toastType = ""
	})
}

// handleGenericEvent is an action handler that receives new SSE events, parses them, and shows notifications.
func (h *Header) handleGenericEvent(ctx app.Context, a app.Action) {
	// Fetch the SSE event.
	//event, ok := a.Value.(sse.Event)
	//event, ok := a.Value.(common.Event)
	ev, ok := a.Value.(app.Value)
	if !ok {
		// Cannot assert the event.
		return
	}

	if ev.Get("eventName").IsNull() || ev.Get("data").IsNull() {
		return
	}
	event := common.Event{Type: ev.Get("eventName").String(), Data: ev.Get("data").String()}

	//fmt.Printf("%s: %s\n", event.Type, event.Data)

	// Exit if the event is a heartbeat. But notice the last timestamp.
	if event.Data == "heartbeat" || event.Type == "keepalive" {
		// Update the timestamp value in the LS.
		ctx.LocalStorage().Set("lastEventTime", time.Now().Unix())

		// Use the content field too for the same action.
		ctx.Dispatch(func(ctx app.Context) {
			h.lastHeartbeatTime = time.Now().Unix()
		})
		return
	}

	// Abort the further stream listening, set time timer for a reconnect.
	/*if event.Type == "close" || event.Type == "server-stop" {
		app.Window().Get(common.JS_LITTR_SSE).Call("abort")

		fmt.Println("SSE client closed, setting a timeout for a reconnection...")
		app.Window().Get(common.JS_LITTR_SSE).Set("reconnection_timeout", 20000)
		app.Window().Get(common.JS_LITTR_SSE).Call("tryReconnect")
	}*/

	// Fetch the user from the LocalStorage.
	var user models.User
	common.LoadUser(&user, &ctx)

	// Do not parse the message when user has live mode disabled.
	/*if !user.LiveMode {
		return
	}*/

	//text, link := event.ParseEventData(&user)

	// Instantiate the new toast.
	//toast := common.Toast{AppContext: &ctx}

	// Show the snack bar the nasty way.
	/*snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() && text != "" {
		snack.Get("classList").Call("add", "active")
		snack.Set("innerHTML", "<a href=\""+link+"\"><i>info</i>"+text+"</a>")
	}*/

	// Change the page's title to indicate a new event present.
	/*title := app.Window().Get("document")
	if !title.IsNull() && !strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", "(*) "+prevTitle)
	}*/

	//toast.Text(text).Link(link).Type(common.TTYPE_INFO).Dispatch(h, dispatch)
}

func (h *Header) handleHeaderClick(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalInfoShow = true
	})
}

func (h *Header) handleInstallClick(ctx app.Context, a app.Action) {
	ctx.ShowAppInstallPrompt()
}

// onKeyDown is a callback to handle the key-down event: this allows one to control the app using their keyboard more effectively.
func (h *Header) handleKeydown(ctx app.Context, a app.Action) {
	event, ok := a.Value.(app.Value)
	if !ok {
		return
	}

	// Was the key Escape/Esc? Then cast general item dismissal.
	if event.Get("key").String() == "Escape" || event.Get("key").String() == "Esc" {
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

	// Prevent sending actions on ctrl+shfit+R.
	if event.Get("ctrlKey").Bool() && event.Get("shiftKey").Bool() && (event.Get("key").String() == "r" || event.Get("key").String() == "R") {
		return
	}

	// Fetch the reply textarea.
	textareaReply := app.Window().GetElementByID("reply-textarea")

	if (event.Get("key").String() == "x" ||
		event.Get("key").String() == "X" ||
		event.Get("key").String() == "r" ||
		event.Get("key").String() == "R") &&
		textareaReply.IsNull() {

		ctx.NewAction("dismiss")
		ctx.NewAction("clear")
		ctx.NewActionWithValue("refresh", event.Get("key").String())
	}

	// Fetch the post textarea.
	textareaPost := app.Window().GetElementByID("post-textarea")

	// Otherwise utilize the CTRL-Enter combination and send the post in.
	if event.Get("ctrlKey").Bool() && event.Get("key").String() == "Enter" {
		// Click the new post button.
		if !textareaPost.IsNull() && len(textareaPost.Get("value").String()) > 0 {
			app.Window().GetElementByID("button-new-post").Call("click")
			return
		}

		// Click the new reply button.
		if !textareaReply.IsNull() && len(textareaReply.Get("value").String()) > 0 {
			app.Window().GetElementByID("button-reply").Call("click")
			return
		}

		win := app.Window()

		// Submit a new poll.
		if !win.GetElementByID("poll-question").IsNull() &&
			len(win.GetElementByID("poll-question").Get("value").String()) > 0 &&
			!win.GetElementByID("poll-option-i").IsNull() &&
			len(win.GetElementByID("poll-option-i").Get("value").String()) > 0 &&
			!win.GetElementByID("poll-option-ii").IsNull() &&
			len(win.GetElementByID("poll-option-ii").Get("value").String()) > 0 &&
			!win.GetElementByID("poll-option-iii").IsNull() &&
			len(win.GetElementByID("poll-option-iii").Get("value").String()) > 0 {

			app.Window().GetElementByID("button-new-poll").Call("click")
		}
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
	switch event.Get("key").String() {
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

func (h *Header) handleLinkClick(ctx app.Context, a app.Action) {
	path := ctx.Page().URL().Path

	switch a.Name {
	case "login-click":
		ctx.Navigate("/login")
	case "stats-click":
		if path != "/stats" {
			ctx.Navigate("/stats")
		}
	case "users-click":
		if path != "/users" {
			ctx.Navigate("/users")
		}
	case "post-click":
		if path != "/post" {
			ctx.Navigate("/post")
		}
	case "polls-click":
		if path != "/polls" {
			ctx.Navigate("/polls")
		}
	case "flow-click":
		if path != "/flow" {
			ctx.Navigate("/flow")
		}
	case "settings-click":
		if path != "/settings" {
			ctx.Navigate("/settings")
		}
	case "user-flow-click":
		id, ok := a.Value.(string)
		if !ok {
			break
		}

		if strings.Contains(ctx.Page().URL().Path, "/flow") {
			ctx.NewAction("dismiss-general")

			if strings.Contains(ctx.Page().URL().Path, id) {
				break
			}
		}

		ctx.Navigate("/flow/users/" + id)
	}
}

func (h *Header) handleLogout(ctx app.Context, _ app.Action) {
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		/*if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}*/

		ctx.Navigate("/logout")
	})
}

func (h *Header) handleReload(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		h.updateAvailable = false
	})

	ctx.LocalStorage().Set("newUpdate", false)
	ctx.Reload()
}

func (h *Header) handleUserModalShow(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		h.modalLogoutShow = true
	})
}
