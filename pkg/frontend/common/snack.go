package common

import (
	//"fmt"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

const (
	GENERIC_TOAST_NAME    = "snackbar-general-bottom"
	GENERIC_TOAST_LINK    = "snackbar-general-bottom-link"
	GENERIC_TOAST_TIMEOUT = 5000
	DISMISS_LOCK          = "dismissLock"
)

func hideGenericToast(toastName, color string, ID int64) {
	if toastName == "" || color == "" {
		return
	}

	toast := app.Window().GetElementByID(toastName)
	if !toast.IsNull() &&
		!toast.Get(DISMISS_LOCK).IsUndefined() &&
		!toast.Get(DISMISS_LOCK).Bool() &&
		!toast.Get("uniqueID").IsUndefined() &&
		int64(toast.Get("uniqueID").Int()) == ID {

		app.Window().GetElementByID(toastName).Get("classList").Call("remove", "active", color)
	}

	// Set the page title's back.
	title := app.Window().Get("document")
	if !title.IsNull() && !toast.Get(DISMISS_LOCK).IsUndefined() && !toast.Get(DISMISS_LOCK).Bool() && strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", prevTitle[4:])
	}
}

type ToastPayload struct {
	// Required ones.
	Name string
	Text string

	// Default: blue10.
	Color string

	// Optionals.
	Link string
	Keep bool
}

// ShowGenericToast is a helper function to show requested toast/snackbar. At the moment, the concept is that there are two nodes loaded in DOM (top and bottom toast/snack) to show common response (top), or system alerts (bottom).
func ShowGenericToast(pl *ToastPayload) {
	if pl.Text == "" || pl.Name == "" {
		return
	}

	toast := app.Window().GetElementByID(pl.Name)
	if !toast.IsNull() {
		var toastClass = func() string {
			if strings.Contains(pl.Name, "top") {
				return "top"
			}

			return "bottom"
		}()

		var toastColor = func() string {
			if pl.Color != "" {
				return pl.Color
			}

			return "blue10"
		}()

		uniqueID := time.Now().Unix()

		app.Window().GetElementByID(pl.Name).Set("className", "")

		// Activate the toast/snackbar. Assign the dismiss lock if requested.
		app.Window().GetElementByID(pl.Name).Get("classList").Call("add", "snackbar", "white-text", toastClass, toastColor, "active")
		app.Window().GetElementByID(pl.Name).Set(DISMISS_LOCK, pl.Keep)
		app.Window().GetElementByID(pl.Name).Set("uniqueID", uniqueID)

		// Set the snackbar's/toast's link.
		if toastLink := app.Window().GetElementByID(pl.Name + "-link"); !toastLink.IsUndefined() && pl.Link != "" {
			toastLink.Set("href", pl.Link)
		}

		// Update the snackbar's/toast's inner HTML content.
		if pl.Keep {
			app.Window().GetElementByID(pl.Name).Set("innerHTML", "<div class=\"max\"><i>info</i>&nbsp;&nbsp;"+pl.Text+"</div><div><i>close</i></div>")
		} else {
			app.Window().GetElementByID(pl.Name).Set("innerHTML", "<div class=\"max\"><i>info</i>&nbsp;&nbsp;"+pl.Text+"</div>")
		}

		if pl.Keep {
			// Change the page's title to indicate a new event present.
			title := app.Window().Get("document")
			if !title.IsNull() && !strings.Contains(title.Get("title").String(), "(*)") {
				prevTitle := title.Get("title").String()
				title.Set("title", "(*) "+prevTitle)
			}
		}

		var timer *time.Timer

		// Register a click event listener.
		app.Window().GetElementByID(pl.Name).Call("addEventListener", "click", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			this.Set(DISMISS_LOCK, false)

			if timer != nil {
				timer.Stop()
			}

			hideGenericToast(pl.Name, toastColor, uniqueID)
			return nil
		}))

		// Hold the toast on mouseover event.
		app.Window().GetElementByID(pl.Name).Call("addEventListener", "mouseenter", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			app.Window().GetElementByID(pl.Name).Set("innerHTML", "<div class=\"max\"><i>info</i>&nbsp;&nbsp;"+pl.Text+"</div><div><i>close</i></div>")
			this.Set(DISMISS_LOCK, true)
			return nil
		}))

		// Handle the timeout of the toast.
		go func() {
			timer = time.NewTimer(time.Millisecond * time.Duration(GENERIC_TOAST_TIMEOUT))

			<-timer.C
			hideGenericToast(pl.Name, toastColor, uniqueID)
		}()
	}
}
