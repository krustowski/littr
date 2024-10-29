package common

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ToastInterface interface {
	Context(*app.Context) *Toast
	Text(string) *Toast
	Link(string) *Toast
	Type(string) *Toast
	SetPrefix(string) *Toast
	RemovePrefix() *Toast
	Dispatch(interface{}, func(*Toast, interface{}))
}

type Toast struct {
	AppContext *app.Context
	TLink      string
	TText      string
	TType      string
	TID        int64
}

const (
	// Toast types.
	TTYPE_ERR     = "error"
	TTYPE_INFO    = "info"
	TTYPE_SUCCESS = "success"
)

var ToastColor = func(ttype string) string {
	switch ttype {
	case TTYPE_SUCCESS:
		return "green10"

	case TTYPE_INFO:
		return "blue10"

	case TTYPE_ERR:
	default:
		return "red10"
	}

	// Set the unknown option to the INFO color.
	return "blue10"
}

func (t *Toast) Context(ctx *app.Context) *Toast {
	t.AppContext = ctx
	return t
}

func (t *Toast) Link(link string) *Toast {
	t.TLink = link
	return t
}

func (t *Toast) Text(text string) *Toast {
	t.TText = text
	return t
}

func (t *Toast) Type(typ string) *Toast {
	t.TType = typ
	return t
}

func (t *Toast) Dispatch(c interface{}, f func(*Toast, interface{})) {
	if f == nil || c == nil {
		return
	}

	id := time.Now().UnixNano()
	t.TID = id

	snack := app.Window().GetElementByID(fmt.Sprintf("snackbar-%d", t.TID))
	if !snack.IsNull() {
		snack.Get("classList").Call("add", "active")
	}

	f(t, c)

	go func(tt *Toast, cc interface{}) {
		time.Sleep(time.Second * 5)

		if tt == nil || cc == nil || tt.TID != id {
			return
		}

		t.Text("")
		f(tt, cc)

		snack := app.Window().GetElementByID(fmt.Sprintf("snackbar-%d", tt.TID))
		if !snack.IsNull() {
			snack.Get("classList").Call("remove", "active")
		}
	}(t, c)
}
