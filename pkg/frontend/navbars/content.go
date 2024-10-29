// The navigation bars (sub)view and view-controllers logic package.
package navbars

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/tmaxmax/go-sse"
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

	// EventListener functions.
	keyDownEventListener   func()
	eventListenerMessage   func()
	eventListenerKeepAlive func()

	// Helper field to catch the timestamp of the last received keepalive event.
	lastHeartbeatTime int64

	// Toast-related fields.
	toast     common.Toast
	toastText string
	toastShow bool
	toastType string

	// Context cancellation function.
	sseCancel context.CancelFunc
}

type Footer struct {
	app.Compo

	authGranted bool
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

	ctx.LocalStorage().Set("newUpdate", true)

	// Force reload the app on update.
	//ctx.Reload()
}

func (h *Header) OnMount(ctx app.Context) {
	// Register the app's indicators.
	h.appInstallable = ctx.IsAppInstallable()
	h.onlineState = true

	// Get the current auth state from LocalStorage.
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	// Redirect client to the unauthorized zone.
	path := ctx.Page().URL().Path
	if !authGranted && path != "/" && path != "/login" && path != "/register" && !strings.Contains(path, "/reset") && path != "/tos" {
		ctx.Navigate("/login")
		return
	}

	// Redirect auth'd client from the unauthorized zone.
	if authGranted && (path == "/" || path == "/login" || path == "/register" || path == "/reset") {
		ctx.Navigate("/flow")
		return
	}

	// Test the Go SSE client implementation.
	// Tests: blocks the client goroutine, therefore no other HTTP request is possible anymore when this implementation is started.
	// Conclusion: must be run in async.
	//common.SSEClient()

	// Create event listener for SSE messages.
	//h.eventListenerMessage = app.Window().AddEventListener("message", h.onMessage)
	//h.eventListenerKeepAlive = app.Window().AddEventListener("keepalive", h.onMessage)
	h.keyDownEventListener = app.Window().AddEventListener("keydown", h.onKeyDown)

	// General action to dismiss all items in the UI.
	ctx.Handle("dismiss-general", h.handleDismiss)
	ctx.Handle("generic-event", h.handleGenericEvent)

	ctx.Dispatch(func(ctx app.Context) {
		h.authGranted = authGranted
		h.pagePath = path
	})

	// Keep the update button on until clicked.
	var newUpdate bool
	ctx.LocalStorage().Get("newUpdate", &newUpdate)

	if newUpdate {
		h.updateAvailable = true
	}

	/*h.onlineState = true // this is a guess
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
	}))*/
}

func (h *Header) OnNav(ctx app.Context) {
	// New context. Notify the context on common syscalls.
	var cctx context.Context
	cctx, h.sseCancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	//defer h.sseCancel()

	// A HTTP request with context.
	req, _ := http.NewRequestWithContext(cctx, http.MethodGet, common.URL+"/api/v1/live", http.NoBody)

	ctx.Async(func() {
		// New SSE connection.
		conn := common.Client.NewConnection(req)

		// Subscribe to any event, regardless the type.
		conn.SubscribeToAll(func(event sse.Event) {
			ctx.NewActionWithValue("generic-event", event)

			fmt.Printf("%s: %s\n", event.Type, event.Data)
			/*switch event.Type {
			case "keepalive", "ops":
				fmt.Printf("%s: %s\n", event.Type, event.Data)
			case "server-stop":
				fmt.Println("server closed!")
				h.sseCancel()
			default: // no event name*/
			//}
		})

		// Create a new connection.
		if err := conn.Connect(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		return
	})
}

func (f *Footer) OnMount(ctx app.Context) {
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	f.authGranted = authGranted
}
