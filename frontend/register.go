package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"log"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"go.savla.dev/littr/backend"
	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

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
		Code  int                    `json:"code"`
		Users map[string]models.User `json:"users"`
	}{}

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		nickname := strings.TrimSpace(c.nickname)
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)
		email := strings.TrimSpace(c.email)

		// fetch the users list to compare to
		resp, ok := litterAPI("GET", "/api/users", nil, nickname, 0)
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

		if nickname == "" || passphrase == "" || passphraseAgain == "" || email == "" {
			toastText = "all fields need to be filled"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// don't allow very long nicknames
		if len(nickname) > config.NicknameLengthMax {
			toastText = "nickname has to be " + strconv.Itoa(config.NicknameLengthMax) + " chars long at max"

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
		for _, user := range response.Users {
			if email != user.Email {
				continue
			}

			toastText = "this email has been already used"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		passHash := sha512.Sum512([]byte(passphrase + config.Pepper))

		var user models.User = models.User{
			Nickname:       nickname,
			Passphrase:     string(passHash[:]),
			Email:          email,
			FlowList:       make(map[string]bool),
			RegisteredTime: time.Now(),
			AvatarURL:      backend.GetGravatarURL(email),
		}

		user.FlowList[nickname] = true
		user.FlowList["system"] = true

		resp, ok = litterAPI("POST", "/api/users", user, user.Nickname, 0)
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

		if response.Code == 409 {
			toastText = "that user already exists!"

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
		app.H5().Text("littr registration").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("do not be mid, join us to be lit"),
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
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("nickname "),
				app.Span().Text("is your unique identifier for the system operations; "),
				app.Span().Text("please double-check your nickname before registering (nickname is case-sensitive)"),
			),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "nickname"),
			app.Label().Text("nickname").Class("active"),
		),

		// password fields
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("passphrase "),
				app.Span().Text("is your secret key to the littr account"),
			),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password"),
			app.Label().Text("passphrase").Class("active"),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password"),
			app.Label().Text("passphrase again").Class("active"),
		),

		// e-mail field
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("e-mail "),
				app.Span().Text("address is used for user's avatar fetching from Gravatar.com, and (not yet implemented) for the account verification, please enter a valid e-mail address"),
			),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("text").OnChange(c.ValueTo(&c.email)).Required(true).Class("active").MaxLength(60).Attr("autocomplete", "email"),
			app.Label().Text("e-mail").Class("active"),
		),

		// GDPR warning
		app.Article().Class("row border").Style("word-break", "break-word").Body(
			app.I().Text("warning"),
			app.Div().Class("max").Style("word-break", "break-word").Style("hyphens", "auto").Body(
				app.P().Style("word-break", "break-word").Style("hyphens", "auto").Body(
					app.Span().Text("by clicking on the register button you are giving us a GDPR consent (a permission to store your account information in the database)"),
				),
				app.P().Text("you can flush your account data and published posts simply on the settings page after a log-in"),
			),
		),

		app.Div().Class("space"),

		// register button
		app.Button().Class("responsive deep-orange7 white-text bold").OnClick(c.onClickRegister).Disabled(c.registerButtonDisabled).Body(
			app.Text("register"),
		),

		app.Div().Class("space"),
	)
}
