package navbars

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

const (
	headerString = "littr"
)

type Header struct {
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

	toast     common.Toast
	toastText string
	toastShow bool
	toastType string
}

type Footer struct {
	app.Compo

	authGranted bool
}

func (h *Header) OnAppUpdate(ctx app.Context) {
	// Reports that an app update is available.
	//h.updateAvailable = ctx.AppUpdateAvailable()

	ctx.Dispatch(func(ctx app.Context) {
		h.updateAvailable = true
	})

	ctx.LocalStorage().Set("newUpdate", true)

	// force reload the app on update
	//ctx.Reload()
}

func (h *Header) OnMount(ctx app.Context) {
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

func (f *Footer) OnMount(ctx app.Context) {
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	f.authGranted = authGranted
}

func (h *Header) OnAppInstallChange(ctx app.Context) {
	ctx.Dispatch(func(ctx app.Context) {
		h.appInstallable = ctx.IsAppInstallable()
	})
}
