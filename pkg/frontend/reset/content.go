package reset

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	email string
	uuid  string

	showUUIDPage bool

	toast common.Toast

	buttonsDisabled bool

	keyDownEventListener func()
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)

	url := strings.Split(ctx.Page().URL().Path, "/")

	toast := common.Toast{AppContext: &ctx}

	// autosend the UUID to the backend if present in URL
	if len(url) > 2 && url[2] != "" {
		uuid := url[2]
		c.showUUIDPage = true

		if err := c.handleResetRequest("", uuid); err != nil {
			toast.Text(err.Error()).Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
			})
			return
		}

		toast.Text("your passphrase has been changed, check your inbox").Type("success").Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			c.buttonsDisabled = false
		})
	}
}