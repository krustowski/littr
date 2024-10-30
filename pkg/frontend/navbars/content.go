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
	"time"

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

	// Context cancellation function for the SSE client.
	sseCancel context.CancelFunc
}

type Footer struct {
	app.Compo

	// Simple authentication indicatior.
	authGranted bool

	// Context cancellation function for the SSE client.
	sseCancel context.CancelFunc
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
}

// Exclussively used for the SSE client as a whole.
func (f *Footer) OnMount(ctx app.Context) {
	var authGranted bool
	ctx.LocalStorage().Get("authGranted", &authGranted)

	f.authGranted = authGranted

	// Do not start the SSE client for the unauthenticated visitors at all.
	if !authGranted {
		return
	}

	// Prepare the variable to load the user's data from LS.
	var user models.User
	common.LoadUser(&user, &ctx)

	// If the options map is nil, or the liveMode is disabled within, do not continue as well.
	if user.Options == nil || (user.Options != nil && !user.Options["liveMode"]) {
		return
	}

	// Custom HTTP client full definition.
	var client = sse.Client{
		// Standard HTTP client.
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				// Idle conn = keeplive conn
				// https://pkg.go.dev/net/http#Transport
				MaxIdleConns:       1,
				IdleConnTimeout:    20 * time.Second,
				DisableCompression: true,
				DisableKeepAlives:  false,
			},
		},
		// Callback function when the connection is to be reastablished.
		OnRetry: func(err error, duration time.Duration) {
			fmt.Printf("retrying: %v\n", err)
			time.Sleep(duration)
		},
		// Validation of the response content-type mainly, e.g. DefaultValidator, or NoopValidator.
		ResponseValidator: common.DefaultValidator,
		// The connection strategy tuning.
		Backoff: sse.Backoff{
			// The initial wait time before a reconnect is attempted.
			InitialInterval: 500 * time.Millisecond,
			// How fast should the reconnection time grow.
			// 1 = constatnt time interval.
			Multiplier: float64(1.5),
			// Jitter: range (0, 1).
			// -1 = no randomization.
			Jitter: float64(0.5),
			// How much can the wait time grow.
			// 0 = grow indefinitely.
			MaxInterval: 2500 * time.Millisecond,
			// Stop retrying after such time.
			// 0 = no limit.
			MaxElapsedTime: 10000 * time.Millisecond,
			// The retry count allowed.
			// 0 = infinite, <0 = no retries.
			MaxRetries: 0,
		},
	}

	ctx.Async(func() {
		//go func() {
		// New context. Notify the context on common syscalls.
		var cctx context.Context
		cctx, f.sseCancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

		defer f.sseCancel()

		// A HTTP request with context.
		req, _ := http.NewRequestWithContext(cctx, http.MethodGet, common.URL+"/api/v1/live", http.NoBody)

		// New SSE connection.
		//conn := common.Client.NewConnection(req)
		conn := client.NewConnection(req)

		// Subscribe to any event, regardless the type.
		conn.SubscribeToAll(func(event sse.Event) {
			ctx.NewActionWithValue("generic-event", event)

			if event.Type == "close" {
				f.sseCancel()
			}

			// Print all events.
			fmt.Printf("%s: %s\n", event.Type, event.Data)
		})

		// Create a new connection.
		if err := conn.Connect(); err != nil {
			fmt.Printf("conn error: %v\n", err)
			//fmt.Fprintln(os.Stderr, err)
		}

		return
	})
	//}()
}

func (f *Footer) OnDismount(ctx app.Context) {
	f.sseCancel()
}
