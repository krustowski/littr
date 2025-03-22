// The navigation bars (sub)view and view-controllers logic package.
package navbars

import (
	"fmt"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	//"github.com/tmaxmax/go-sse"
)

const (
	headerString = "littr"
)

type Header struct {
	app.Compo

	// Boolean app's indicators.
	updateAvailable bool
	appInstallable  bool

	// Auth&user-related fields.
	authGranted bool
	user        models.User

	// Modal fields.
	modalInfoShow   bool
	modalLogoutShow bool

	// Experimental function.
	onlineState bool

	pagePath string

	lastHeartbeatTime int64

	// Toast-related fields.
	toastTop    common.Toast
	toastBottom common.Toast

	toastText string
	toastShow bool
	toastType string

	sseConnStatus string
}

type Footer struct {
	app.Compo

	// Simple authentication indicatior.
	authGranted bool

	user models.User
}

func (h *Header) OnAppInstallChange(ctx app.Context) {
	ctx.Dispatch(func(ctx app.Context) {
		h.appInstallable = ctx.IsAppInstallable()
	})
}

func (h *Header) OnAppUpdate(ctx app.Context) {
	// Reports that an app update is available.
	//h.updateAvailable = ctx.AppUpdateAvailable()

	ctx.Dispatch(func(ctx app.Context) {
		h.updateAvailable = true
	})

	ctx.SetState(common.StateNameNewUpdate, true)
}

func (h *Header) OnMount(ctx app.Context) {
	if app.IsServer {
		return
	}

	// Register the app's indicators.
	h.appInstallable = ctx.IsAppInstallable()
	h.onlineState = true

	// Keep the update button on until clicked.
	var newUpdate bool
	ctx.GetState(common.StateNameNewUpdate, &newUpdate)

	if newUpdate {
		h.updateAvailable = true
	}

	//
	//  Auth-based navigations
	//

	// Get the current auth state from LocalStorage.
	var authGranted bool
	ctx.GetState(common.StateNameAuthGranted, &authGranted)

	// Redirect client to the unauthorized zone.
	path := ctx.Page().URL().Path
	if !authGranted && path != "/" && path != "/login" && path != "/register" && !strings.Contains(path, "/reset") && !strings.Contains(path, "/success") && path != "/tos" {
		ctx.Navigate("/login")
		return
	}

	// Redirect auth'd client from the unauthorized zone.
	if authGranted && (path == "/" || path == "/login" || path == "/register" || path == "/reset") {
		ctx.Navigate("/flow")
		return
	}

	ctx.GetState(common.StateNameUser, &h.user)
	h.ensureUIColors()

	// Custom SSE client implementation.
	if !app.Window().Get(common.JsLittrSse).Get("running").Bool() {
		fmt.Println("Connecting to the SSE stream...")
		app.Window().Get(common.JsLittrSse).Call("tryReconnect")
	}

	//
	//  Event Listeners
	//

	var addTrackedEventListener = app.FuncOf(func(this app.Value, args []app.Value) any {
		const eventRegistry = "eventRegistry"

		if len(args) < 3 {
			return nil
		}

		element := args[0]
		eventType := args[1]
		listener := args[2]

		if element.Type() != app.TypeObject || eventType.Type() != app.TypeString || listener.Type() != app.TypeFunction {
			return nil
		}

		registry := app.Window().Get(eventRegistry)

		if registry.IsNull() || registry.IsUndefined() {
			app.Window().Set(eventRegistry, app.Window().Get("Array").New())
			registry = app.Window().Get(eventRegistry)
		}

		for i := 0; i < registry.Length(); i++ {
			elem := registry.Index(i)

			if elem.Get("element").Equal(element) && elem.Get("eventType").Equal(eventType) {
				return nil
			}
		}

		fmt.Println("adding new eventListener:", eventType.String())

		registry.Call("push", map[string]interface{}{
			"element":   element,
			"eventType": eventType,
			"listener":  listener,
		})

		element.Call("addEventListener", eventType, listener)

		return nil
	})

	if app.Window().Get("addTrackedEventListener").IsNull() || app.Window().Get("addTrackedEventListener").IsUndefined() {
		app.Window().Set("addTrackedEventListener", addTrackedEventListener)
	}

	var onlineHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		tPl := &common.ToastPayload{
			Name:  "snackbar-general-bottom",
			Text:  common.MSG_STATE_ONLINE,
			Link:  "",
			Color: "blue10",
			Keep:  false,
		}

		common.ShowGenericToast(tPl)
		return nil
	})

	var offlineHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		tPl := &common.ToastPayload{
			Name:  "snackbar-general-bottom",
			Text:  common.MSG_STATE_OFFLINE,
			Link:  "",
			Color: "blue10",
			Keep:  true,
		}

		common.ShowGenericToast(tPl)
		return nil
	})

	var keydownHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		ctx.NewActionWithValue("keydown", args[0])
		return nil
	})

	var scrollHandler = app.FuncOf(func(this app.Value, args []app.Value) any {
		ctx.NewAction("scroll")
		return nil
	})

	app.Window().Call("addTrackedEventListener", app.Window(), "online", onlineHandler)
	app.Window().Call("addTrackedEventListener", app.Window(), "offline", offlineHandler)
	app.Window().Call("addTrackedEventListener", app.Window(), "keydown", keydownHandler)
	app.Window().Call("addTrackedEventListener", app.Window(), "scroll", scrollHandler)

	//
	//  Action handlers
	//

	// General action to dismiss all items in the UI.
	ctx.Handle("dismiss-general", h.handleDismiss)
	//ctx.Handle("generic-event", h.handleGenericEvent)
	ctx.Handle("keydown", h.handleKeydown)
	ctx.Handle("littr-header-click", h.handleHeaderClick)
	ctx.Handle("reload", h.handleReload)
	ctx.Handle("user-modal-show", h.handleUserModalShow)
	ctx.Handle("install-click", h.handleInstallClick)
	ctx.Handle("logout", h.handleLogout)

	ctx.Handle("login-click", h.handleLinkClick)
	ctx.Handle("stats-click", h.handleLinkClick)
	ctx.Handle("users-click", h.handleLinkClick)
	ctx.Handle("post-click", h.handleLinkClick)
	ctx.Handle("polls-click", h.handleLinkClick)
	ctx.Handle("flow-click", h.handleLinkClick)
	ctx.Handle("settings-click", h.handleLinkClick)
	ctx.Handle("user-flow-click", h.handleLinkClick)

	ctx.Dispatch(func(ctx app.Context) {
		h.authGranted = authGranted
		h.pagePath = path
	})
}

func (f *Footer) OnNav(ctx app.Context) {
	if app.IsServer {
		return
	}

	// Prepare the variable to load the user's data from LS.
	if err := common.LoadUser(&f.user, &ctx); err != nil {
		return
	}
}

// Exclussively used for the SSE client as a whole.
func (f *Footer) OnMount(ctx app.Context) {
	if app.IsServer {
		return
	}

	ctx.GetState(common.StateNameAuthGranted, &f.authGranted)
	ctx.GetState(common.StateNameUser, &f.user)
}
