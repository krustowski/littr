package polls

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Toast struct {
	AppContext *app.Context
	ToastLink  string
	ToastText  string
	ToastType  string
}

func (t *Toast) Context(ctx *app.Context) *Toast {
	t.AppContext = ctx
	return t
}

func (t *Toast) Link(link string) *Toast {
	t.ToastLink = link
	return t
}

func (t *Toast) Text(text string) *Toast {
	t.ToastText = text
	return t
}

func (t *Toast) Type(typ string) *Toast {
	t.ToastType = typ
	return t
}

func (t *Toast) Dispatch(c *Content) {
	if t.AppContext == nil {
		return
	}

	(*t.AppContext).Dispatch(func(ctx app.Context) {
		c.toastText = t.ToastText
		c.toastShow = (t.ToastText != "")
		c.toastType = t.ToastType
		//c.toastLink = t.Link

		c.toast = *t
	})
	return
}
