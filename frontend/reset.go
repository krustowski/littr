package frontend

import (
	"encoding/json"
	"log"
	"strconv"
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

	email string

	toastShow bool
	toastText string
	toastType string

	buttonsDisabled bool
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
		Message     string                 `json:"message"`
		AuthGranted bool                   `json:"auth_granted"`
		Users       map[string]models.User `json:"users"`
		Code        int                    `json:"code"`
		//FlowRecords []string `json:"flow_records"`
	}{}
	toastText := ""

	c.buttonsDisabled = true

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

		respRaw, ok := litterAPI("POST", "/api/auth/password", &models.User{
			Nickname:   "",
			Passphrase: "",
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

			rr := *respRaw
			log.Println(string(rr[:]))

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if response.Code != 200 {
			code := strconv.Itoa(response.Code)
			toastText = "backend error: code " + code

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
	c.buttonsDisabled = false
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
		app.H5().Text("littr passphrase reset").Style("padding-top", config.HeaderTopPadding),
		app.P().Body(
			app.P().Text("actual pwd is about to be yeeted"),
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
			app.Input().Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.email)).Class("active").Attr("autocomplete", ""),
			app.Label().Text("email").Class("active deep-orange-text"),
		),

		app.Div().Class("small-space"),

		// pwd reset button
		app.Button().Class("responsive deep-orange7 white-text bold").TabIndex(3).OnClick(c.onClick).Disabled(c.buttonsDisabled).Body(
			app.Text("reset"),
		),
		app.Div().Class("space"),
	)
}
