package navbars

import (
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/tmaxmax/go-sse"
)

func (h *Header) handleDismiss(ctx app.Context, a app.Action) {
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
		h.toast.TText = ""
	})
}

// handleGenericEvent is an action handler that receives new SSE events, parses them, and shows notifications.
func (h *Header) handleGenericEvent(ctx app.Context, a app.Action) {
	event, ok := a.Value.(sse.Event)
	if !ok {
		// Cannot assert the event.
		return
	}

	// Exit if the event is a heartbeat. But notice the last timestamp.
	if event.Data == "heartbeat" || event.Type == "keepalive" {
		ctx.Dispatch(func(ctx app.Context) {
			h.lastHeartbeatTime = time.Now().Unix()
		})
		return
	}

	var baseString string

	// Fetch the user from the LocalStorage.
	ctx.LocalStorage().Get("user", &baseString)
	user := common.LoadUser(baseString)

	// Do not parse the message when user has live mode disabled.
	if !user.LiveMode {
		return
	}

	// Explode the data CSV string.
	slice := strings.Split(event.Data, ",")
	text := ""

	switch slice[0] {
	// Server is stopping, being stopped, restarting etc.
	case "server-stop":
		text = "server is restarting..."
		break

	// Server is booting up (just started).
	case "server-start":
		text = "server has just started"
		break

	// New post added.
	case "post":
		author := slice[1]
		if author == user.Nickname {
			return
		}

		// Exit when the author is not followed, nor is found in the user's flowList.
		if flowed, found := user.FlowList[author]; !flowed || !found {
			return
		}

		// Notify the user via toast.
		text = "new post added by " + author
		break

	// New poll added.
	case "poll":
		text = "new poll has been added"
		break
	}

	// Show the snack bar the nasty way.
	snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() && text != "" {
		snack.Get("classList").Call("add", "active")
		snack.Set("innerHTML", "<i>info</i>"+text)
	}

	// Change the title to indicate a new event.
	title := app.Window().Get("document")
	if !title.IsNull() && !strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", "(*) "+prevTitle)
	}

	// Won't trigger the render for some reason... see the bypass ^^
	/*ctx.Dispatch(func(ctx app.Context) {
		//h.toastText = "new post added above"
		//h.toastType = "info"
	})*/
	return
}
