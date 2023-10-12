package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"strings"

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

	loginButtonDisabled bool
}

func (p *LoginPage) OnMount(ctx app.Context) {
	if ctx.Page().URL().Path == "/logout" {
		// destroy auth manually without API
		//ctx.LocalStorage().Set("userLogged", false)
		//ctx.LocalStorage().Set("userName", "")
		//ctx.LocalStorage().Set("flowRecords", nil)
		ctx.LocalStorage().Set("user", "")

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
	return
}

func (c *loginContent) onClick(ctx app.Context, e app.Event) {
	response := struct {
		Message     string `json:"message"`
		AuthGranted bool   `json:"auth_granted"`
		//FlowRecords []string `json:"flow_records"`
		Users map[string]models.User `json:"users"`
	}{}
	toastText := ""

	// fix this!
	c.loginButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)

		if nickname == "" || passphrase == "" {
			toastText = "all fields need to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		passHash := sha512.Sum512([]byte(passphrase + config.Pepper))

		respRaw, ok := litterAPI("POST", "/api/auth", &models.User{
			Nickname:   nickname,
			Passphrase: string(passHash[:]),
		}, nickname)

		if !ok {
			toastText = "backend error: API call failed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if respRaw == nil {
			toastText = "backend error: blank response from API"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if err := json.Unmarshal(*respRaw, &response); err != nil {
			toastText = "backend error: cannot unmarshal response: " + err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if !response.AuthGranted {
			toastText = "wrong credentials on input"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		user, err := json.Marshal(response.Users[nickname])
		if err != nil {
			toastText = "frontend error: user marshal failed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// save enrypted user data to their Local browser storage
		ctx.LocalStorage().Set("user", config.Encrypt(config.Pepper, string(user)))

		if response.AuthGranted {
			ctx.Navigate("/flow")
		}
	})

}

func (c *loginContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = false
	c.loginButtonDisabled = false
}

func (c *loginContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr login").Style("padding-top", config.HeaderTopPadding),
		app.P().Body(
			app.P().Text("littr, bc even litter can be lit"),
		),
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// login credentials fields
		app.Div().Class("field border label invalid deep-orange-text").Body(
			app.Input().Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(config.NicknameLengthMax).Class("active"),
			app.Label().Text("nickname").Class("active"),
		),

		app.Div().Class("field border label invalid deep-orange-text").Body(
			app.Input().Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Class("active").AutoComplete(true),
			app.Label().Text("passphrase").Class("active"),
		),

		// login button
		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).OnClick(c.onClick).Disabled(c.loginButtonDisabled).Body(
			app.Text("login"),
		),
		app.Div().Class("space"),

		app.P().Class("center-align").Text("or"),
		app.Div().Class("space"),

		// register button
		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).OnClick(c.onClickRegister).Disabled(c.loginButtonDisabled).Body(
			app.Text("register"),
		),

		app.Div().Class("space"),
	)
}
