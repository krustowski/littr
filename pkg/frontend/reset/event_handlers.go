package reset

import (
	"net/mail"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) onClickRequest(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.buttonsDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		email := strings.TrimSpace(c.email)

		if email == "" && !app.Window().GetElementByID("email-input").IsNull() {
			email = strings.TrimSpace(app.Window().GetElementByID("email-input").Get("value").String())
		}

		if email == "" {
			toast.Text("e-mail field has to be filled").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
			})
			return
		}

		// validate e-mail struct
		// https://stackoverflow.com/a/66624104
		if _, err := mail.ParseAddress(email); err != nil {
			toast.Text("wrong e-mail format entered").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
			})
			return
		}

		if err := c.handleResetRequest(email, ""); err != nil {
			toast.Text(err.Error()).Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
			})
			return
		}

		toast.Text("passphrase reset request sent successfully, check your inbox").Type("success").Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			c.showUUIDPage = true
			c.buttonsDisabled = false
		})
		return
	})
}

func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	emailInput := app.Window().GetElementByID("email-input")
	uuidInput := app.Window().GetElementByID("uuid-input")

	if (emailInput.IsNull() && !c.showUUIDPage) || (uuidInput.IsNull() && c.showUUIDPage) {
		return
	}

	if !emailInput.IsNull() && len(emailInput.Get("value").String()) == 0 && !c.showUUIDPage {
		return
	}

	if !uuidInput.IsNull() && len(uuidInput.Get("value").String()) == 0 && c.showUUIDPage {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" {
		if c.showUUIDPage {
			if !app.Window().GetElementByID("reset-button").IsNull() {
				app.Window().GetElementByID("reset-button").Call("click")
			}
		} else {
			if !app.Window().GetElementByID("request-button").IsNull() {
				app.Window().GetElementByID("request-button").Call("click")
			}
		}
	}
}

func (c *Content) onClickReset(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	c.buttonsDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		uuid := strings.TrimSpace(c.uuid)

		if uuid == "" && !app.Window().GetElementByID("uuid-input").IsNull() {
			uuid = strings.TrimSpace(app.Window().GetElementByID("uuid-input").Get("value").String())
		}

		if uuid == "" {
			toast.Text("please insert UUID from your inbox").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
			})
			return
		}

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
		return
	})
}
