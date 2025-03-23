package login

import (
	"regexp"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) onClick(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	ctx.Dispatch(func(ctx app.Context) {
		c.loginButtonDisabled = true
	})

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.loginButtonDisabled = false
		})

		// Trim the padding spaces on the extremities.
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)

		if nickname == "" && !app.Window().GetElementByID("login-input").IsNull() {
			nickname = strings.TrimSpace(app.Window().GetElementByID("login-input").Get("value").String())
		}

		if passphrase == "" && !app.Window().GetElementByID("passphrase-input").IsNull() {
			passphrase = strings.TrimSpace(app.Window().GetElementByID("passphrase-input").Get("value").String())
		}

		if nickname == "" || passphrase == "" {
			toast.Text(common.ERR_ALL_FIELDS_REQUIRED).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Don't allow special chars and spaces in the nickname
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toast.Text(common.ERR_LOGIN_CHARS_LIMIT).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		payload := &struct {
			Nickname        string `json:"nickname"`
			PassphrasePlain string `json:"passphrase_plain"`
		}{
			Nickname:        nickname,
			PassphrasePlain: passphrase,
		}

		input := &common.CallInput{
			Method: "POST",
			Url:    "/api/v1/auth",
			Data:   payload,
		}

		type dataModel struct {
			AuthGranted bool         `json:"auth_granted"`
			User        *models.User `json:"user"`
		}

		output := &common.Response{Data: &dataModel{}}

		if ok := common.FetchData(input, output); !ok {
			if output.Error != nil {
				toast.Text(output.Error.Error()).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if !data.AuthGranted {
			toast.Text(common.ERR_ACCESS_DENIED).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		ctx.SetState(common.StateNameAuthGranted, true).Persist()
		ctx.SetState(common.StateNameUser, data.User).Persist()

		if data.AuthGranted {
			ctx.Navigate("/flow")
		}
	})

}

func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	loginInput := app.Window().GetElementByID("login-input")
	passphraseInput := app.Window().GetElementByID("passphrase-input")

	if loginInput.IsNull() || passphraseInput.IsNull() {
		return
	}

	if len(loginInput.Get("value").String()) == 0 || len(passphraseInput.Get("value").String()) == 0 {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" {
		app.Window().GetElementByID("login-button").Call("click")
	}
}

func (c *Content) onClickRegister(ctx app.Context, _ app.Event) {
	ctx.Navigate("/register")
}

func (c *Content) onClickReset(ctx app.Context, _ app.Event) {
	ctx.Navigate("/reset")
}
