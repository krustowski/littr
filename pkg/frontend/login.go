package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/models"

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

	keyDownEventListener func()
}

func (p *LoginPage) OnMount(ctx app.Context) {
	if ctx.Page().URL().Path == "/logout" {
		// destroy auth manually without API
		//ctx.LocalStorage().Set("userLogged", false)
		//ctx.LocalStorage().Set("userName", "")
		//ctx.LocalStorage().Set("flowRecords", nil)
		ctx.SetState("user", "")
		ctx.SetState("authGranted", false)

		p.userLogged = false

		ctx.Navigate("/login")
	}
}

func (p *LoginPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("login / littr")
}

func (p *LoginPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&loginContent{},
	)
}

func (c *loginContent) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)

	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}

func (c *loginContent) onClickRegister(ctx app.Context, e app.Event) {
	ctx.Navigate("/register")
	return
}

func (c *loginContent) onClickReset(ctx app.Context, e app.Event) {
	ctx.Navigate("/reset")
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

		if passphrase == "" && !app.Window().GetElementByID("passphrase-input").IsNull() {
			passphrase = strings.TrimSpace(app.Window().GetElementByID("passphrase-input").Get("value").String())
		}

		if nickname == "" || passphrase == "" {
			toastText = "all fields need to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
			return
		}

		// don't allow special chars and spaces in the nickname
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toastText = "nickname can contain only chars a-z, A-Z and numbers"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
			return
		}

		//passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		passHash := sha512.Sum512([]byte(passphrase + appPepper))

		respRaw, ok := littrAPI("POST", "/api/v1/auth/", &models.User{
			Nickname:      nickname,
			Passphrase:    string(passHash[:]),
			PassphraseHex: fmt.Sprintf("%x", passHash),
		}, nickname, 0)

		if !ok {
			toastText = "backend error: API call failed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
			return
		}

		if respRaw == nil {
			toastText = "backend error: blank response from API"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
			return
		}

		if err := json.Unmarshal(*respRaw, &response); err != nil {
			toastText = "backend error: cannot unmarshal response: " + err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
			return
		}

		if !response.AuthGranted {
			toastText = "wrong credentials passed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
			return
		}

		user, err := json.Marshal(response.Users[nickname])
		if err != nil {
			toastText = "frontend error: user marshal failed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.loginButtonDisabled = false
			})
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

func (c *loginContent) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = ""
		c.toastShow = false
		c.loginButtonDisabled = false
	})
}

func (c *loginContent) dismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *loginContent) onKeyDown(ctx app.Context, e app.Event) {
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

func (c *loginContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("littr login"),
			),
		),
		/*app.P().Body(
			app.P().Text("littr, bc even littr can be lit"),
		),*/
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().ID("snackbar").Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// login credentials fields
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("login-input").Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(configs.NicknameLengthMax).Class("active").Attr("autocomplete", "username"),
			app.Label().Text("nickname").Class("active deep-orange-text"),
		),

		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("passphrase-input").Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Class("active").Attr("autocomplete", "current-password"),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text(" "),
				app.Span().Text("log-in for 30 days"),
			),
		),
		app.Div().Class("space"),

		// login button
		app.Div().Class("row center-align").Body(
			app.Button().ID("login-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(3).OnClick(c.onClick).Disabled(c.loginButtonDisabled).Body(
				app.Text("login"),
			),
		),
		app.Div().Class("space"),

		// reset button
		app.Div().Class("row center-align").Body(
			app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(4).OnClick(c.onClickReset).Disabled(c.loginButtonDisabled).Body(
				app.Text("recover forgotten passphrase"),
			),
		),
		app.Div().Class("space"),

		// register button
		app.Div().Class("row center-align").Body(
			// register button
			app.If(app.Getenv("REGISTRATION_ENABLED") == "true",
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(5).OnClick(c.onClickRegister).Disabled(c.loginButtonDisabled).Body(
					app.Text("register"),
				),
			).Else(
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(5).OnClick(nil).Disabled(true).Body(
					app.Text("register"),
				),
			),
		),
		app.Div().Class("medium-space"),
	)
}
