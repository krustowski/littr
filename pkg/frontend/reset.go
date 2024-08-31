package frontend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"strings"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type ResetPage struct {
	app.Compo
	userLogged bool
}

type resetContent struct {
	app.Compo

	email string
	uuid  string

	showUUIDPage bool

	toastShow bool
	toastText string
	toastType string

	buttonsDisabled bool

	keyDownEventListener func()
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

	respRaw, ok := littrAPI("POST", "/api/v1/users/passphrase/"+path, payload, "", 0)
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

	return nil
}

func (c *resetContent) onClickReset(ctx app.Context, e app.Event) {
	toastText := ""

	c.buttonsDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		uuid := strings.TrimSpace(c.uuid)

		if uuid == "" && !app.Window().GetElementByID("uuid-input").IsNull() {
			uuid = strings.TrimSpace(app.Window().GetElementByID("uuid-input").Get("value").String())
		}

		if uuid == "" {
			toastText = "please insert UUID from your inbox"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.buttonsDisabled = false
			})
			return
		}

		if err := c.handleResetRequest("", uuid); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = true
				c.toastText = err.Error()
				c.toastShow = (toastText != "")
				c.buttonsDisabled = false
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "your passphrase has been changed, check your inbox"
			c.toastType = "success"
			c.toastShow = (toastText != "")
			c.buttonsDisabled = false
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

		if email == "" && !app.Window().GetElementByID("email-input").IsNull() {
			email = strings.TrimSpace(app.Window().GetElementByID("email-input").Get("value").String())
		}

		if email == "" {
			toastText = "e-mail field has to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.buttonsDisabled = false
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
				c.buttonsDisabled = false
			})
			return
		}

		if err := c.handleResetRequest(email, ""); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = true
				c.toastText = err.Error()
				c.toastShow = (toastText != "")
				c.buttonsDisabled = false
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = "passphrase reset request sent successfully, check your inbox"
			c.toastType = "success"
			c.toastShow = (toastText != "")
			c.showUUIDPage = true
			c.buttonsDisabled = false
		})
		return
	})
}

func (c *resetContent) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = ""
		c.toastShow = false
		c.buttonsDisabled = false
	})
}

func (c *resetContent) dismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *resetContent) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	emailInput := app.Window().GetElementByID("email-input")
	uuidInput := app.Window().GetElementByID("uuid-input")

	if (emailInput.IsNull() && !c.showUUIDPage) || (uuidInput.IsNull() && c.showUUIDPage) {
		return
	}

	if !emailInput.IsNull() && len(emailInput.Get("value").String()) == 0 && !c.showUUIDPage {
		return
	}

	if !uuidInput.IsNull() && len(uuidInput.Get("value").String()) == 0 && c.showUUIDPage {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" {
		if c.showUUIDPage {
			app.Window().GetElementByID("reset-button").Call("click")
		} else {
			app.Window().GetElementByID("request-button").Call("click")
		}
	}
}

func (c *resetContent) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)

	url := strings.Split(ctx.Page().URL().Path, "/")

	// autosend the UUID to the backend if present in URL
	if len(url) > 2 && url[2] != "" {
		uuid := url[2]
		c.showUUIDPage = true

		if err := c.handleResetRequest("", uuid); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonsDisabled = false
				c.toastText = err.Error()
				c.toastShow = true
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.buttonsDisabled = false
			c.toastText = "your passphrase has been changed, check your inbox"
			c.toastType = "success"
			c.toastShow = true
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
				app.If(!c.showUUIDPage,
					app.H5().Text("littr passphrase request"),
				).Else(
					app.H5().Text("littr passphrase reset"),
				),
			),
		),

		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().ID("snackbar").Class("snackbar "+toastColor+" white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// passphrase request --- insert an e-mail
		app.If(!c.showUUIDPage,

			// pwd reset lightbulb
			app.Article().Class("row surface-container-highest").Body(
				app.I().Text("lightbulb").Class("amber-text"),
				app.P().Class("max").Body(
					//app.Span().Class("deep-orange-text").Text(" "),
					app.Span().Text("to request a passphrase change, enter your registration e-mail address below, which is linked with your account; a confirmation mail will then be sent to your inbox"),
				),
			),
			app.Div().Class("space"),

			// pwd reset credentials fields
			app.Div().Class("field border label deep-orange-text").Body(
				app.Input().ID("email-input").Type("email").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.email)).Class("active").Attr("autocomplete", "email").AutoFocus(true),
				app.Label().Text("e-mail").Class("active deep-orange-text"),
			),

			//app.Div().Class("small-space"),

			// request button
			app.Div().Class("row center-align").Body(
				app.Button().ID("request-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickRequest).Disabled(c.buttonsDisabled).TabIndex(2).Body(
					app.Text("request"),
				),
			),

		// passphrase reset --- insert the UUID
		).Else(

			// pwd reset lightbulb
			app.Article().Class("row surface-container-highest").Body(
				app.I().Text("lightbulb").Class("amber-text"),
				app.P().Class("max").Body(
					//app.Span().Class("deep-orange-text").Text(" "),
					app.Span().Text("enter the UUID code which has been sent to your inbox; if the code is correct, your passphrase will be automatically regenerated and another confirmation mail containing the passphrase will be sent to your e-mail address"),
				),
			),
			app.Div().Class("space"),

			// pwd reset credentials fields
			app.Div().Class("field border label deep-orange-text").Body(
				app.Input().ID("uuid-input").Type("text").Required(true).TabIndex(1).Value("").OnChange(c.ValueTo(&c.uuid)).Class("active").AutoFocus(true),
				app.Label().Text("UUID").Class("active deep-orange-text"),
			),

			//app.Div().Class("small-space"),

			// pwd reset button
			app.Div().Class("row center-align").Body(
				app.Button().ID("reset-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(2).OnClick(c.onClickReset).Disabled(c.buttonsDisabled).Body(
					app.Text("reset"),
				),
			),
		),

		app.Div().Class("space"),
	)
}
