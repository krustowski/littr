package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type RegisterPage struct {
	app.Compo
}

type registerContent struct {
	app.Compo

	toastShow bool
	toastText string

	users map[string]models.User

	nickname        string
	passphrase      string
	passphraseAgain string
	email           string

	registerButtonDisabled bool

	keyDownEventListener func()
}

func (p *RegisterPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("register / littr")
}

func (p *RegisterPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&registerContent{},
	)
}

func (c *registerContent) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)

	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}

func (c *registerContent) onClickRegister(ctx app.Context, e app.Event) {
	c.registerButtonDisabled = true
	toastText := ""

	response := struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Users   map[string]models.User `json:"users"`
	}{}

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		email := strings.TrimSpace(c.email)

		if email == "" {
			email = strings.TrimSpace(app.Window().GetElementByID("email-input").Get("value").String())
		}

		// fetch the users list to compare to
		/*resp, ok := littrAPI("GET", "/api/users", nil, nickname, 0)
		if !ok {
			toastText = "cannot send API request (backend error)"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if err := json.Unmarshal(*resp, &response); err != nil {
			toastText = "cannot unmarshal response"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}*/

		if nickname == "" || passphrase == "" || passphraseAgain == "" || email == "" {
			toastText = "all fields need to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
			})
			return
		}

		// don't allow very long nicknames
		if len(nickname) > configs.NicknameLengthMax {
			toastText = "nickname has to be " + strconv.Itoa(configs.NicknameLengthMax) + " chars long at max"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
			})
			return
		}

		// don't allow special chars and spaces in the nickname
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toastText = "nickname can contain only chars a-z, A-Z and numbers"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
			})
			return
		}

		// do passphrases match?
		if passphrase != passphraseAgain {
			toastText = "passphrases don't match!"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
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
				c.registerButtonDisabled = false
			})
			return
		}

		// check if the e-mail address has been used already
		/*for _, user := range response.Users {
			if email != user.Email {
				continue
			}

			toastText = "this email has been already used"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}*/

		//passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		passHash := sha512.Sum512([]byte(passphrase + appPepper))

		var user models.User = models.User{
			Nickname:       nickname,
			PassphraseHex:  fmt.Sprintf("%x", passHash),
			Email:          email,
			FlowList:       make(map[string]bool),
			RegisteredTime: time.Now(),
			AvatarURL:      db.GetGravatarURL(email, nil),
		}

		user.FlowList[nickname] = true
		user.FlowList["system"] = true

		input := callInput{
			Method: "POST",
			Url: "/api/v1/users",
			Data: user,
			CallerID: user.Nickname,
			PageNo: 0,
			HideReplies: false,
		}

		resp, ok := littrAPI(input)
		if !ok {
			toastText = "cannot send API request (backend error)"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
			})
			return
		}

		if err := json.Unmarshal(*resp, &response); err != nil {
			toastText = "cannot unmarshal response"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
			})
			return
		}

		if response.Code != 201 {
			//toastText = "that user already exists!"
			toastText = response.Message

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.registerButtonDisabled = false
			})
			return
		}

		if toastText == "" {
			ctx.Navigate("/login")
		}

	})
}

func (c *registerContent) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = ""
		c.toastShow = false
		c.registerButtonDisabled = false
	})
}

func (c *registerContent) dismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *registerContent) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	nicknameInput := app.Window().GetElementByID("nickname-input")
	passphraseInput := app.Window().GetElementByID("passphrase-input")
	passphraseAgainInput := app.Window().GetElementByID("passphrase-again-input")
	emailInput := app.Window().GetElementByID("email-input")

	if nicknameInput.IsNull() || passphraseInput.IsNull() || passphraseAgainInput.IsNull() || emailInput.IsNull() {
		return
	}

	if len(nicknameInput.Get("value").String()) == 0 || len(passphraseInput.Get("value").String()) == 0 || len(passphraseAgainInput.Get("value").String()) == 0 || len(emailInput.Get("value").String()) == 0 {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" {
		app.Window().GetElementByID("register-button").Call("click")
	}
}

func (c *registerContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("littr registration"),
				//app.P().Text("do not be mid, join us to be lit"),
			),
		),
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

		// nickname field
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("nickname "),
				app.Span().Text("is your unique identifier for the system operations; "),
				app.Span().Text("please double-check your nickname before registering (nickname is case-sensitive)"),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("nickname-input").Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true).Class("active").AutoFocus(true).MaxLength(50).Attr("autocomplete", "username").TabIndex(1).Name("login"),
			app.Label().Text("nickname").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// password fields
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("passphrase "),
				app.Span().Text("is your secret key to the littr account"),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-input").Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(2),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-again-input").Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(3),
			app.Label().Text("passphrase again").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// e-mail field
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("e-mail "),
				app.Span().Text("address is used for user's avatar fetching from Gravatar.com, and (not yet implemented) for the account verification, please enter a valid e-mail address"),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("email-input").Type("email").OnChange(c.ValueTo(&c.email)).Required(true).Class("active").MaxLength(60).Attr("autocomplete", "email").TabIndex(4),
			app.Label().Text("e-mail").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// GDPR warning
		app.Article().Class("row surface-container-highest").Style("word-break", "break-word").Body(
			app.I().Text("warning").Class("red-text"),
			app.Div().Class("max").Style("word-break", "break-word").Style("hyphens", "auto").Body(
				app.P().Style("word-break", "break-word").Style("hyphens", "auto").Body(
					app.Span().Text("by clicking on the register button you are giving us a GDPR consent (a permission to store your account information in the database)"),
				),
				app.P().Text("you can flush your account data and published posts simply on the settings page after a log-in"),
			),
		),
		app.Div().Class("space"),

		// register button
		app.Div().Class("row center-align").Body(
			app.If(app.Getenv("REGISTRATION_ENABLED") == "true",
				app.Button().ID("register-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickRegister).Disabled(c.registerButtonDisabled).TabIndex(5).Body(
					app.Text("register"),
				),
			).Else(
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(nil).Disabled(true).Body(
					app.Text("registration off"),
				),
			),
		),

		app.Div().Class("medium-space"),
	)
}
