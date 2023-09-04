package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"os"

	"litter-go/backend"
	"litter-go/config"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type LoginPage struct {
	app.Compo
	userLogged bool
}

type loginContent struct {
	app.Compo

	nickname   string
	passphrase string

	toastShow bool
	toastText string
}

func (p *LoginPage) OnMount(ctx app.Context) {
	if ctx.Page().URL().Path == "/logout" {
		// destroy auth manually without API
		ctx.LocalStorage().Set("userLogged", false)
		ctx.LocalStorage().Set("userName", "")
		ctx.LocalStorage().Set("flowRecords", nil)

		p.userLogged = false

		ctx.Navigate("/login")
	}
}

func (p *LoginPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("login / littr")
}

func (p *LoginPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&loginContent{},
	)
}

func (c *loginContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		c.toastShow = true
		if c.nickname == "" || c.passphrase == "" {
			c.toastText = "all fields need to be filled"
			return
		}

		response := struct {
			Message     string   `json:"message"`
			AuthGranted bool     `json:"auth_granted"`
			FlowRecords []string `json:"flow_records"`
		}{}

		passHash := sha512.Sum512([]byte(c.passphrase + os.Getenv("APP_PEPPER")))

		respRaw, _ := litterAPI("POST", "/api/auth", &backend.User{
			Nickname:   c.nickname,
			Passphrase: string(passHash[:]),
		})

		if respRaw == nil {
			c.toastText = "generic backend error"
			return
		}

		_ = json.Unmarshal(*respRaw, &response)
		if !response.AuthGranted {
			//c.toastText = response.Message
			c.toastText = "access denied"
			return
		}

		c.toastShow = false
		ctx.LocalStorage().Set("userLogged", true)
		ctx.LocalStorage().Set("userName", c.nickname)
		ctx.LocalStorage().Set("flowRecords", response.FlowRecords)
		ctx.Navigate("/flow")
	})
}

func (c *loginContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *loginContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr login").Style("padding-top", config.HeaderTopPadding),
		app.P().Body(
			app.A().Href("/register").Text("littr, bc even litter can be lit ---> register here"),
		),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)),
			app.Label().Text("nickname"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)),
			app.Label().Text("passphrase"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).Text("login").OnClick(c.onClick),
	)
}
