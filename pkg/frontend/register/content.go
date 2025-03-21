// The register view and view-controllers logic package.
package register

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	toast common.Toast

	nickname        string
	passphrase      string
	passphraseAgain string
	email           string

	registerButtonDisabled bool
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)
}
