package frontend

import (
	"crypto/sha512"
	"encoding/json"
	//"log"
	//"os"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

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

func (c *loginContent) onClickRegister(ctx app.Context, e app.Event) {
	ctx.Navigate("/register")
}

func (c *loginContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.nickname == "" || c.passphrase == "" {
			c.toastShow = true
			c.toastText = "all fields need to be filled"
			return
		}

		passHash := sha512.Sum512([]byte(c.passphrase + config.Pepper))

		respRaw, ok := litterAPI("POST", "/api/auth", &models.User{
			Nickname:   c.nickname,
			Passphrase: string(passHash[:]),
		})

		if !ok {
			c.toastShow = true
			c.toastText = "backend error: API call failed"
			return
		}

		if respRaw == nil {
			c.toastShow = true
			c.toastText = "backend error: blank response from API"
			return
		}

		response := struct {
			Message     string `json:"message"`
			AuthGranted bool   `json:"auth_granted"`
			//FlowRecords []string `json:"flow_records"`
			Users map[string]models.User `json:"users"`
		}{}

		if err := json.Unmarshal(*respRaw, &response); err != nil {
			c.toastShow = true
			c.toastText = "backend error: cannot unmarshal response: " + err.Error()
			return
		}

		if !response.AuthGranted {
			c.toastShow = true
			c.toastText = "access denied"
			return
		}

		c.toastShow = false
		ctx.LocalStorage().Set("userLogged", true)
		ctx.LocalStorage().Set("userName", c.nickname)

		user, err := json.Marshal(response.Users[c.nickname])
		if err != nil {
			c.toastShow = true
			c.toastText = "frontend error: user marshal failed"
			return
		}

		ctx.LocalStorage().Set("user", config.Encrypt(config.Pepper, string(user)))

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
			app.P().Text("littr, bc even litter can be lit"),
		),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field border invalid deep-orange-text").Body(
			app.Input().Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(30).Min(5).Class("active"),
			app.Label().Text("nickname").Class("active"),
		),
		app.Div().Class("field border invalid deep-orange-text").Body(
			app.Input().Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Min(8).Class("active").AutoComplete(true),
			app.Label().Text("passphrase").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).Text("login").OnClick(c.onClick),
		app.Div().Class("space"),
		app.P().Class("center-align").Text("or"),
		app.Div().Class("space"),

		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).Text("register").OnClick(c.onClickRegister),
	)
}
