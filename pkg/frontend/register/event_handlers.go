package register

import (
	"crypto/sha512"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// onClickRegister is a callback function to perform a user registration precedure.
func (c *Content) onClickRegister(ctx app.Context, e app.Event) {
	// Very nasty way of disabling the buttons.
	c.registerButtonDisabled = true

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.registerButtonDisabled = false
		})

		// Trim the padding spaces on the extremities.
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		email := strings.TrimSpace(c.email)

		// Try to get the e-mail address once again using JS.
		if email == "" {
			email = strings.TrimSpace(app.Window().GetElementByID("email-input").Get("value").String())
		}

		// Validate that every of these strings are not empty.
		if nickname == "" || passphrase == "" || passphraseAgain == "" || email == "" {
			toast.Text(common.ERR_REGISTER_FIELDS_REQUIRED).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Don't allow very long nicknames (nor the short ones.
		if len(nickname) > config.NicknameLengthMax || len(nickname) < 3 {
			toast.Text("Nickname has to be " + strconv.Itoa(config.NicknameLengthMax) + " chars long at max, or at least 3 characters long.").Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Don't allow special chars and spaces in the nickname.
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toast.Text(common.ERR_REGISTER_CHARSET_LIMIT).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Ensure both passphrases match.
		if passphrase != passphraseAgain {
			toast.Text(common.ERR_PASSPHRASE_MISMATCH).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Validate the e-mail struct.
		// https://stackoverflow.com/a/66624104
		if _, err := mail.ParseAddress(email); err != nil {
			toast.Text(common.ERR_WRONG_EMAIL_FORMAT).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Prepare the passphrase to be encoded into a hexadecimal representation of a sha512sum hash.
		passHash := sha512.Sum512([]byte(passphrase + common.AppPepper))

		// Compose the new User payload.
		var user models.User = models.User{
			Nickname:       nickname,
			PassphraseHex:  fmt.Sprintf("%x", passHash),
			Email:          email,
			FlowList:       make(map[string]bool),
			RegisteredTime: time.Now(),
		}

		// Fetch the user's Gravatar according to their e-mail address.
		//avatarURL := db.GetGravatarURL(user, nil, nil, nil)
		//user.AvatarURL = avatarURL

		// Set default flowList items.
		user.FlowList[nickname] = true
		user.FlowList["system"] = true

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/users",
			Data:        user,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare a blank API response object.
		output := &common.Response{}

		// Execute the user registration procedure.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Check for the HTTP 201 response code, otherwise print the API response message in the toast.
		if output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// If the toast text is empty, continue to the login page.
		if toast.TText == "" {
			ctx.Navigate("/success/registration")
		}
	})
}

// onDismissToast is a callback function to cast a new valueless dismiss action.
func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

// onKeyDown is a callback function to handle the keyboard keys utilization for the UI controlling.
func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	// If the pressed key was Escape/Esc, cast a new dismiss action.
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	// Define the JS objects to be tested for values.
	nicknameInput := app.Window().GetElementByID("nickname-input")
	passphraseInput := app.Window().GetElementByID("passphrase-input")
	passphraseAgainInput := app.Window().GetElementByID("passphrase-again-input")
	emailInput := app.Window().GetElementByID("email-input")

	// Exit if any of the JS object is nil.
	if nicknameInput.IsNull() || passphraseInput.IsNull() || passphraseAgainInput.IsNull() || emailInput.IsNull() {
		return
	}

	// Exit if any of the JS object's value is zero.
	if len(nicknameInput.Get("value").String()) == 0 || len(passphraseInput.Get("value").String()) == 0 || len(passphraseAgainInput.Get("value").String()) == 0 || len(emailInput.Get("value").String()) == 0 {
		return
	}

	// Click the register button if everything passes.
	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" {
		app.Window().GetElementByID("register-button").Call("click")
	}
}
