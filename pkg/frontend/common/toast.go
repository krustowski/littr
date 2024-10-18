package common

import (
	"fmt"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ToastInterface interface {
	Dispatch() func(interface{})
	False() func(interface{})
}

type Toast struct {
	AppContext *app.Context
	TLink      string
	TText      string
	TType      string
	TID        int64
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
