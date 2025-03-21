// The reset view and view-controllers logic package.
package reset

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	email string
	uuid  string

	showUUIDPage bool

	toast common.Toast

	buttonsDisabled bool
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)
	//c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)

	url := strings.Split(ctx.Page().URL().Path, "/")

	toast := common.Toast{AppContext: &ctx}

	// autosend the UUID to the backend if present in URL
	if len(url) > 2 && url[2] != "" {
		uuid := url[2]
		c.showUUIDPage = true

		if err := c.handleResetRequest("", uuid); err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
			})
			return
		}

		toast.Text(common.MSG_RESET_PASSPHRASE_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			c.buttonsDisabled = false
		})
	}
}
