package login

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	nickname   string
	passphrase string

	toast common.Toast

	loginButtonDisabled bool

	keyDownEventListener func()
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)

	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}
