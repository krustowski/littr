package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"strings"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ResetPage struct {
	app.Compo
	userLogged bool
}

type resetContent struct {
	app.Compo

	email   string

	toastShow bool
	toastText string

	buttonDisabled bool
}

func (p *ResetPage) OnMount(ctx app.Context) {
}

func (p *ResetPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("reset / littr")
}

func (p *ResetPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&resetContent{},
	)
}

func (c *resetContent) onClick(ctx app.Context, e app.Event) {
	response := struct {
		Message     string `json:"message"`
		AuthGranted bool   `json:"auth_granted"`
		//FlowRecords []string `json:"flow_records"`
		Users map[string]models.User `json:"users"`
	}{}
	toastText := ""

	c.buttonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		email := strings.TrimSpace(c.email)

		if email == "" {
			toastText = "e-mail field has to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		password :=  RandStringBytesMaskImprSrc(16)

		passHash := sha512.Sum512([]byte(password + config.Pepper))

		respRaw, ok := litterAPI("POST", "/api/user/password", &models.User{
			Nickname:   "",
			Passphrase: string(passHash[:]),
			Email:      email,
		}, "", 0)

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

		if response.Code != 200 {
			code := strings.Itoa(response.Code)
			toastText = "backend error: code "+code

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		/*if toastText == "" {
			ctx.Navigate("/login")
		}*/

			toastText = "reset e-mail sent"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastType = "success"
				c.toastShow = (toastText != "")
			})
			return
	})

}

func (c *resetContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = false
	c.buttonDisabled = false
}

func (c *resetContent) Render() app.UI {
	toastColor := ""

	switch c.toastType {
	case "success":
		toastColor = "green10"
		break

	case "info":
		toastColor = "blue10"
		break

	default:
		toastColor = "red10"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr login").Style("padding-top", config.HeaderTopPadding),
		app.P().Body(
			app.P().Text("littr, bc even litter can be lit"),
		),
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar "+toastColor+" white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// pwd reset lightbulb
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text(" "),
				app.Span().Text("enter your e-mail address below; after that, password of the linked account will be reset, and a confirmation mail will be sent to such address if found"),
			),
		),

		// pwd reset credentials fields
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(config.NicknameLengthMax).Class("active").Attr("autocomplete", "nickname"),
			app.Label().Text("nickname").Class("active deep-orange-text"),
		),

		app.Div().Class("small-space"),

		// pwd reset button
		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).OnClick(c.onClick).Disabled(c.buttonsDisabled).Body(
			app.Text("reset"),
		),
		app.Div().Class("space"),
	)
}
