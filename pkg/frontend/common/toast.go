package common

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ToastInterface interface {
	// Context method sets the application context pointer reference.
	Context(*app.Context) *Toast

	// Text method write the input string to the TText field.
	Text(string) *Toast

	// Link method writes the input string to the TLink field.
	Link(string) *Toast

	// Type method writes the input string to the TType field.
	Type(string) *Toast

	// SetPrefix method enables to set the logging prefix. Can be removed afterwards.
	SetPrefix(string) *Toast

	// RemovePrefix method removes the previously added logging prefix.
	RemovePrefix() *Toast

	// Dispatch sends the instance itself to the Content type of a view. This method is the final one.
	Dispatch(interface{}, func(*Toast, interface{}))
}

type Toast struct {
	// AppContext is a pointer reference to the application context.
	AppContext *app.Context

	// TLink is a field to hold the hypertext link.
	TLink string

	// TText is a filed to hold the very text message to display.
	TText string

	// TType defines the message type (error, info, success).
	TType string

	// TID is a filed to hold the toast's UUID.
	TID int64
}

const (
	// Toast type error = red10.
	TTYPE_ERR = "error"

	// Toast type info = blue10.
	TTYPE_INFO = "info"

	// Toast type success = green10.
	TTYPE_SUCCESS = "success"
)

// ToastColor is a helper function reference to define the colour palette for such toast types.
var ToastColor = func(ttype string) string {
	switch ttype {
	// Type success.
	case TTYPE_SUCCESS:
		return "green10"

	// Type error.
	case TTYPE_ERR:
		return "red10"

	// Type info.
	case TTYPE_INFO:
	default:
		return "blue10"
	}

	// Set the unknown option to the INFO color.
	return "blue10"
}

// Context sets the application context pointer reference. Returns itself.
func (t *Toast) Context(ctx *app.Context) *Toast {
	t.AppContext = ctx
	return t
}

// Link sets the string input as the TLink content. Returns itself.
func (t *Toast) Link(link string) *Toast {
	t.TLink = link
	return t
}

// Text sets the string input as the TText content. Returns itself.
func (t *Toast) Text(text string) *Toast {
	t.TText = text
	return t
}

// Type sets the string input as the TType content. Returns itself.
func (t *Toast) Type(typ string) *Toast {
	t.TType = typ
	return t
}

// Dispatch is the final method for the toast's cycle. This method ensures a proper propagation of the toast to such screen to display its content.
// Custom implementations of the <f> function can be seens in other packages that use this very implementation.
func (t *Toast) Dispatch(c interface{}, f func(*Toast, interface{})) {
	// If the function and/or Content interface are nil, exit.
	if f == nil || c == nil {
		return
	}

	// Compose the toast ID and assign it.
	id := time.Now().UnixNano()
	t.TID = id

	// Fetch the generic toast and rewrite its ID to match the just-obtained one.
	snack := app.Window().GetElementByID(fmt.Sprintf("snackbar-%d", t.TID))
	// If the object seems to exist, make it active = visible.
	if !snack.IsNull() {
		snack.Get("classList").Call("add", "active")
	}

	// Run the custom dispatch implementation.
	f(t, c)

	// Start a async goroutine, pass the toast pointer and content interface in.
	go func(tt *Toast, cc interface{}) {
		// Set the timeout for such toast.
		time.Sleep(time.Second * 5)

		// Verify the values of inputs.
		if tt == nil || cc == nil || tt.TID != id {
			return
		}

		// Empty the text contents.
		t.Text("")

		// Re-run the custom dispatch implementation.
		f(tt, cc)

		// Fetch the toast with the previosly rewitten ID.
		snack := app.Window().GetElementByID(fmt.Sprintf("snackbar-%d", tt.TID))
		// If it exists, make it not.
		if !snack.IsNull() {
			snack.Get("classList").Call("remove", "active")
		}
	}(t, c)
}
