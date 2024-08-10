package frontend

import (
	"encoding/json"
	"fmt"
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

func (c *resetContent) handleResetRequest(email, uuid string) error {
	if email == "" && uuid == "" {
		return fmt.Errorf("invalid payload data")
	}

	path := "request"

	if uuid != "" {
		path = "reset"
	}

	payload := struct {
		Email string `json:"email"`
		UUID  string `json:"uuid"`
	}{
		Email: email,
		UUID:  uuid,
	}

	respRaw, ok := litterAPI("POST", "/api/v1/users/passphrase/"+path, payload, "", 0)

	if !ok {
		return fmt.Errorf("communication with backend failed")
	}

	if respRaw == nil {
		return fmt.Errorf("no data received from backend")
	}

	response := struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{}

	if err := json.Unmarshal(*respRaw, &response); err != nil {
		log.Println(err.Error())
		return fmt.Errorf("corrupted data received from backend")
	}

	if response.Code != 200 {
		return fmt.Errorf("%s", response.Message)
	}
}

func (c *resetContent) onClickReset(ctx app.Context, e app.Event) {
	toastText := ""

	c.buttonsDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		uuid := strings.TrimSpace(c.UUID)

		if uuid == "" {
			toastText = "please insert UUID from your inbox"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if err := handleResetRequest("", uuid); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = true
				c.toastText = err.Error()
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.buttonsDisabled = true
			c.toastText = "your passphrase has been changed, check your inbox"
			c.toastType = "success"
			c.toastShow = (toastText != "")
		})
		return
	})
}

func (c *resetContent) onClickRequest(ctx app.Context, e app.Event) {
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

		if err := handleResetRequest(email, ""); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = true
				c.toastText = err.Error()
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "passphrase reset request sent successfully, check your inbox"
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

func (c *resetContent) OnMount(ctx app.Context) {
	url := strings.Split(ctx.Page().URL().Path, "/")

	// autosend the UUID to the backend if present in URL
	if len(url) > 2 && url[2] != "" {
		c.UUID = url[2]

		if err := handleResetRequest("", uuid); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = true
				c.toastText = err.Error()
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.buttonsDisabled = true
			c.toastText = "your passphrase has been changed, check your inbox"
			c.toastType = "success"
			c.toastShow = (toastText != "")
		})
	}
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
			app.Input().Type("email").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.email)).Class("active").Attr("autocomplete", "email").AutoFocus(true).TabIndex(1),
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
