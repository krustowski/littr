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

	"go.savla.dev/littr/configs"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/models"

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

		// fetch the users list to compare to
		/*resp, ok := litterAPI("GET", "/api/users", nil, nickname, 0)
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
			})
			return
		}

		// don't allow very long nicknames
		if len(nickname) > configs.NicknameLengthMax {
			toastText = "nickname has to be " + strconv.Itoa(configs.NicknameLengthMax) + " chars long at max"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// don't allow special chars and spaces in the nickname
		if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nickname) {
			toastText = "nickname can contain only chars a-z, A-Z and numbers"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// do passphrases match?
		if passphrase != passphraseAgain {
			toastText = "passphrases don't match!"

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

		resp, ok := litterAPI("POST", "/api/v1/users", user, user.Nickname, 0)
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
		}

		if response.Code != 201 {
			//toastText = "that user already exists!"
			toastText = response.Message

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if toastText == "" {
			ctx.Navigate("/login")
		}

	})
}

func (c *registerContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = false
	c.registerButtonDisabled = false
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
				app.Div().Class("snackbar red10 white-text top active").Body(
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
			app.Input().Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true).Class("active").AutoFocus(true).MaxLength(50).Attr("autocomplete", "nickname").TabIndex(1),
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
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(2),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(3),
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
			app.Input().Type("email").OnChange(c.ValueTo(&c.email)).Required(true).Class("active").MaxLength(60).Attr("autocomplete", "email").TabIndex(4),
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
		app.Div().Class("medium-space"),

		// register button
		app.Div().Class("row").Body(
			app.If(app.Getenv("REGISTRATION_ENABLED") == "true",
				app.Button().Class("max deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickRegister).Disabled(c.registerButtonDisabled).TabIndex(5).Body(
					app.Text("register"),
				),
			).Else(
				app.Button().Class("max deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(nil).Disabled(true).Body(
					app.Text("registration off"),
				),
			),
		),

		app.Div().Class("space"),
	)
}
