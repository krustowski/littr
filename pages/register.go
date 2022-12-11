package pages

import (
	"crypto/sha512"
	"litter-go/backend"
	"log"
	"net/mail"
	"os"

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
		if c.nickname == "" || c.passphrase == "" || c.passphraseAgain == "" || c.email == "" {
			c.toastShow = true
			c.toastText = "all fields need to be filled"
			return
		}

		if len(c.nickname) > 20 {
			c.toastShow = true
			c.toastText = "nickname has to be 20 chars long at max"
			return
		}

		if c.passphrase != c.passphraseAgain {
			c.toastShow = true
			c.toastText = "passphrases don't match!"
			return
		}

		// validate e-mail struct
		// https://stackoverflow.com/a/66624104
		if _, err := mail.ParseAddress(c.email); err != nil {
			log.Println(err)
			c.toastShow = true
			c.toastText = "wrong e-mail format entered"
			return
		}

		passHash := sha512.Sum512([]byte(c.passphrase + os.Getenv("APP_PEPPER")))

		var user backend.User = backend.User{
			Nickname:   c.nickname,
			Passphrase: string(passHash[:]),
			Email:      c.email,
		}

		if _, ok := litterAPI("POST", "/api/users", user); !ok {
			c.toastShow = true
			c.toastText = "cannot send API request (backend error)"
			return
		}

		c.toastShow = false
		ctx.Navigate("/login")
	})
}

func (c *registerContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *registerContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr registration"),
		app.P().Text("do not be mid, join us to be lit"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true),
			app.Label().Text("nickname"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true),
			app.Label().Text("passphrase again"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.email)).Required(true),
			app.Label().Text("e-mail"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("register").OnClick(c.onClick),
	)
}
