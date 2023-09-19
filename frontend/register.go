package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"log"
	"net/mail"
	"strings"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type RegisterPage struct {
	app.Compo
}

type registerContent struct {
	app.Compo

	toastShow bool
	toastText string

	nickname        string
	passphrase      string
	passphraseAgain string
	email           string
}

func (p *RegisterPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("register / littr")
}

func (p *RegisterPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&registerContent{},
	)
}

func (c *registerContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		email := strings.TrimSpace(c.email)

		if nickname == "" || passphrase == "" || passphraseAgain == "" || email == "" {
			c.toastText = "all fields need to be filled"
			return
		}

		// don't allow very long nicknames
		if len(nickname) > 20 {
			c.toastText = "nickname has to be 20 chars long at max"
			return
		}

		// do passphrases match?
		if passphrase != passphraseAgain {
			c.toastText = "passphrases don't match!"
			return
		}

		// validate e-mail struct
		// https://stackoverflow.com/a/66624104
		if _, err := mail.ParseAddress(email); err != nil {
			log.Println(err)
			c.toastText = "wrong e-mail format entered"
			return
		}

		passHash := sha512.Sum512([]byte(passphrase + config.Pepper))

		var user models.User = models.User{
			Nickname:   nickname,
			Passphrase: string(passHash[:]),
			Email:      email,
			FlowList:   make(map[string]bool),
		}
		user.FlowList[nickname] = true
		user.FlowList["system"] = true

		resp, ok := litterAPI("POST", "/api/users", user)
		if !ok {
			c.toastText = "cannot send API request (backend error)"
			return
		}

		response := struct {
			Code int `json:"code"`
		}{}

		if err := json.Unmarshal(*resp, &response); err != nil {
			c.toastText = "cannot unmarshal response"
			return
		}

		if response.Code == 409 {
			c.toastText = "that user already exists!"
			return
		}

		// update the context of UI goroutine 
		ctx.Dispatch(func(ctx app.Context) {
			c.toastShow = (c.toastText != "")
	
			if !c.toastShow {
				ctx.Navigate("/login")
			}
		})
	})
}

func (c *registerContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = false
}

func (c *registerContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastText != "" {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr registration").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("do not be mid, join us to be lit"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true).Class("active").MaxLength(50),
			app.Label().Text("nickname").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true).Class("active").MaxLength(50).AutoComplete(true),
			app.Label().Text("passphrase").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true).Class("active").MaxLength(50).AutoComplete(true),
			app.Label().Text("passphrase again").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.email)).Required(true).Class("active").MaxLength(60),
			app.Label().Text("e-mail").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("register").OnClick(c.onClick).Disabled(false),
	)
}
