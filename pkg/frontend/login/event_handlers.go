package login

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) onClick(ctx app.Context, e app.Event) {
	toast := common.Toast{AppContext: &ctx}

	// nasty
	c.loginButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)

		if passphrase == "" && !app.Window().GetElementByID("passphrase-input").IsNull() {
			passphrase = strings.TrimSpace(app.Window().GetElementByID("passphrase-input").Get("value").String())
		}

		if nickname == "" || passphrase == "" {
			toast.Text("all fields need to be filled").Type("error").Dispatch(c, dispatch)
			return
		}

		// don't allow special chars and spaces in the nickname
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toast.Text("nickname can contain only chars a-z, A-Z and numbers").Type("error").Dispatch(c, dispatch)
			return
		}

		//passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		passHash := sha512.Sum512([]byte(passphrase + common.AppPepper))

		payload := &models.User{
			Nickname:      nickname,
			Passphrase:    string(passHash[:]),
			PassphraseHex: fmt.Sprintf("%x", passHash),
		}

		input := common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/auth/",
			Data:        payload,
			CallerID:    nickname,
			PageNo:      0,
			HideReplies: false,
		}

		response := struct {
			Message     string `json:"message"`
			AuthGranted bool   `json:"auth_granted"`
			//FlowRecords []string `json:"flow_records"`
			Users map[string]models.User `json:"users"`
		}{}

		if ok := common.CallAPI(input, &response); !ok {
			toast.Text("backend error: API call failed").Type("error").Dispatch(c, dispatch)
			return
		}

		if !response.AuthGranted {
			toast.Text("wrong credentials passed").Type("error").Dispatch(c, dispatch)
			return
		}

		user, err := json.Marshal(response.Users[nickname])
		if err != nil {
			toast.Text("frontend error: user marshal failed").Type("error").Dispatch(c, dispatch)
			return
		}

		// save enrypted user data to their Local browser storage
		ctx.LocalStorage().Set("user", user)
		ctx.LocalStorage().Set("authGranted", true)

		if response.AuthGranted {
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

func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *Content) onClickRegister(ctx app.Context, e app.Event) {
	ctx.Navigate("/register")
	return
}

func (c *Content) onClickReset(ctx app.Context, e app.Event) {
	ctx.Navigate("/reset")
	return
}
