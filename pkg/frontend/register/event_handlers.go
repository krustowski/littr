package register

import (
	"crypto/sha512"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) onClickRegister(ctx app.Context, e app.Event) {
	c.registerButtonDisabled = true
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		email := strings.TrimSpace(c.email)

		if email == "" {
			email = strings.TrimSpace(app.Window().GetElementByID("email-input").Get("value").String())
		}

		if nickname == "" || passphrase == "" || passphraseAgain == "" || email == "" {
			toast.Text("all fields need to be filled").Type("error").Dispatch(c, dispatch)
			return
		}

		// don't allow very long nicknames
		if len(nickname) > configs.NicknameLengthMax {
			toast.Text("nickname has to be "+strconv.Itoa(configs.NicknameLengthMax)+" chars long at max").Type("error").Dispatch(c, dispatch)
			return
		}

		// don't allow special chars and spaces in the nickname
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toast.Text("nickname can contain only chars a-z, A-Z and numbers").Type("error").Dispatch(c, dispatch)
			return
		}

		// do passphrases match?
		if passphrase != passphraseAgain {
			toast.Text("passphrases don't match!").Type("error").Dispatch(c, dispatch)
			return
		}

		// validate e-mail struct
		// https://stackoverflow.com/a/66624104
		if _, err := mail.ParseAddress(email); err != nil {
			toast.Text("wrong e-mail format entered").Type("error").Dispatch(c, dispatch)
			return
		}

		passHash := sha512.Sum512([]byte(passphrase + common.AppPepper))

		var user models.User = models.User{
			Nickname:       nickname,
			PassphraseHex:  fmt.Sprintf("%x", passHash),
			Email:          email,
			FlowList:       make(map[string]bool),
			RegisteredTime: time.Now(),
			AvatarURL:      db.GetGravatarURL(email, nil),
		}

		user.FlowList[nickname] = true
		user.FlowList["system"] = true

		input := common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/users",
			Data:        user,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		response := struct {
			Code    int                    `json:"code"`
			Message string                 `json:"message"`
			Users   map[string]models.User `json:"users"`
		}{}

		if ok := common.CallAPI(input, &response); !ok {
			toast.Text("cannot send API request (backend error)").Type("error").Dispatch(c, dispatch)
			return
		}

		// has the user been registred?
		if response.Code != 201 {
			toast.Text(response.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		if toast.TText == "" {
			ctx.Navigate("/login")
		}

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

	nicknameInput := app.Window().GetElementByID("nickname-input")
	passphraseInput := app.Window().GetElementByID("passphrase-input")
	passphraseAgainInput := app.Window().GetElementByID("passphrase-again-input")
	emailInput := app.Window().GetElementByID("email-input")

	if nicknameInput.IsNull() || passphraseInput.IsNull() || passphraseAgainInput.IsNull() || emailInput.IsNull() {
		return
	}

	if len(nicknameInput.Get("value").String()) == 0 || len(passphraseInput.Get("value").String()) == 0 || len(passphraseAgainInput.Get("value").String()) == 0 || len(emailInput.Get("value").String()) == 0 {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" {
		app.Window().GetElementByID("register-button").Call("click")
	}
}
