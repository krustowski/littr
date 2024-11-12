package common

import (
	//"fmt"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

const (
	GENERIC_TOAST_NAME    = "snackbar-general"
	GENERIC_TOAST_LINK    = "snackbar-general-link"
	GENERIC_TOAST_TIMEOUT = 5000
	DISMISS_LOCK          = "dismissLock"
)

func hideGenericToast() {
	toast := app.Window().GetElementByID(GENERIC_TOAST_NAME)
	if !toast.IsNull() && !toast.Get(DISMISS_LOCK).IsUndefined() && !toast.Get(DISMISS_LOCK).Bool() {
		//app.Window().GetElementByID(GENERIC_TOAST_NAME).Call("removeEventListener", "mouseenter")
		//app.Window().GetElementByID(GENERIC_TOAST_NAME).Call("removeEventListener", "click")
		app.Window().GetElementByID(GENERIC_TOAST_NAME).Get("classList").Call("remove", "active")
	}

	// Set the page title's back.
	title := app.Window().Get("document")
	if !title.IsNull() && !toast.Get(DISMISS_LOCK).IsUndefined() && !toast.Get(DISMISS_LOCK).Bool() && strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", prevTitle[4:])
	}
}

func ShowGenericToast(text, link string, keep bool) {
	if text == "" {
		//hideGenericToast()
		return
	}

	toast := app.Window().GetElementByID(GENERIC_TOAST_NAME)
	if !toast.IsNull() {
		app.Window().GetElementByID(GENERIC_TOAST_NAME).Get("classList").Call("add", "active")
		app.Window().GetElementByID(GENERIC_TOAST_NAME).Set(DISMISS_LOCK, keep)

		// Set the snackbar's/toast's link.
		if toastLink := app.Window().GetElementByID(GENERIC_TOAST_LINK); !toastLink.IsUndefined() && link != "" {
			toastLink.Set("href", link)
		}

		// Update the snackbar's/toast's inner HTML content.
		if keep {
			app.Window().GetElementByID(GENERIC_TOAST_NAME).Set("innerHTML", "<div class=\"max\"><i>info</i>&nbsp;"+text+"</div><div>(close)</div>")
		} else {
			app.Window().GetElementByID(GENERIC_TOAST_NAME).Set("innerHTML", "<div class=\"max\"><i>info</i>&nbsp;"+text+"</div>")
		}

		if keep {
			// Change the page's title to indicate a new event present.
			title := app.Window().Get("document")
			if !title.IsNull() && !strings.Contains(title.Get("title").String(), "(*)") {
				prevTitle := title.Get("title").String()
				title.Set("title", "(*) "+prevTitle)
			}
		}

		// Register a click event listener.
		app.Window().GetElementByID(GENERIC_TOAST_NAME).Call("addEventListener", "click", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			this.Set(DISMISS_LOCK, false)
			hideGenericToast()
			return nil
		}))

		// Hold the toast on mouseover event.
		app.Window().GetElementByID(GENERIC_TOAST_NAME).Call("addEventListener", "mouseenter", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
			this.Set(DISMISS_LOCK, true)
			return nil
		}))

		var timer *time.Timer

		// Handle the timeout of the toast.
		go func() {
			timer = time.NewTimer(time.Millisecond * time.Duration(GENERIC_TOAST_TIMEOUT))

			<-timer.C
			hideGenericToast()
		}()
	}
}
