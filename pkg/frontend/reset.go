package frontend

import (
	"encoding/json"
	"log"
	"net/mail"
	"strconv"
	"strings"

	"go.savla.dev/littr/pkg/models"

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

		// validate e-mail struct
		// https://stackoverflow.com/a/66624104
		if _, err := mail.ParseAddress(email); err != nil {
			log.Println(err)
			toastText = "wrong e-mail format entered"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		respRaw, ok := litterAPI("PATCH", "/api/v1/users/passphrase", &models.User{
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
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("littr passphrase reset"),
			),
		),
		/*app.P().Body(
			app.P().Text("actual pwd is about to be yeeted"),
		),*/
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
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text(" "),
				app.Span().Text("enter your e-mail address below; after that, password of the linked account will be reset, and a confirmation mail will be sent to such address if found"),
			),
		),
		app.Div().Class("space"),

		// pwd reset credentials fields
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().Type("email").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.email)).Class("active").AutoComplete(true).AutoFocus(true).TabIndex(1),
			app.Label().Text("e-mail").Class("active deep-orange-text"),
		),

		//app.Div().Class("small-space"),

		// pwd reset button
		app.Div().Class("row").Body(
			app.Button().Class("max deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(1).OnClick(c.onClick).Disabled(c.buttonsDisabled).TabIndex(2).Body(
				app.Text("reset"),
			),
		),
		app.Div().Class("space"),
	)
}
