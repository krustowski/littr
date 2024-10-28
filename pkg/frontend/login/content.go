package login

import (
	"strings"

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

	activationUUID string
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)

	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}

func (c *Content) OnNav(ctx app.Context) {
	url := strings.Split(ctx.Page().URL().Path, "/")

	var activationUUID string

	// Look if we got the right path format and content = parse the URL.
	if len(url) > 2 && url[2] != "" {
		switch url[1] {
		case "activation":
			activationUUID = url[2]
		}
	} else {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.activationUUID = activationUUID
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// Compose the API call input.
		input := &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/users/activation/" + activationUUID,
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		// Call the API to fetch the data.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Something went wrong...
		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// User successfully activated.
		toast.Text(common.MSG_USER_ACTIVATED).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)
		return
	})
}
