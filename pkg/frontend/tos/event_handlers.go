package tos

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) onClickDismiss(ctx app.Context, e app.Event) {
	c.toastShow = false
	c.toast.TText = ""
	//c.buttonDisabled = false
}
