// The register view and view-controllers logic package.
package register

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	toast common.Toast

	users map[string]models.User

	nickname        string
	passphrase      string
	passphraseAgain string
	email           string

	registerButtonDisabled bool

	keyDownEventListener func()
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)

	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}
