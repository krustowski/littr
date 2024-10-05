package common

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ToastInterface interface {
	Dispatch() func(interface{})
}

type Toast struct {
	AppContext *app.Context
	TLink      string
	TText      string
	TType      string
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

	f(t, c)
}
