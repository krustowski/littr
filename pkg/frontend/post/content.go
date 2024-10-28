// The post (new flow post, or new poll) view and view-controllers logic package.
package post

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	postType string

	newPost    string
	newFigLink string
	newFigFile string
	newFigData []byte

	pollQuestion  string
	pollOptionI   string
	pollOptionII  string
	pollOptionIII string

	toast common.Toast

	postButtonsDisabled bool

	keyDownEventListener func()
}

func (c *Content) OnMount(ctx app.Context) {
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
	ctx.Handle("dismiss", c.handleDismiss)
}
