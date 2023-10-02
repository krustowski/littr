package frontend

import (
	"crypto/sha512"
	"strings"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type SettingsPage struct {
	app.Compo
}

type settingsContent struct {
	app.Compo

	// TODO: review this
	loggedUser string

	// used with forms
	passphrase      string
	passphraseAgain string
	aboutText       string

	// loaded logged user's struct
	user models.User

	// message toast vars
	toastShow bool
	toastText string

	settingsButtonDisabled bool
}

func (p *SettingsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&settingsContent{},
	)
}

func (p *SettingsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("settings / littr")
}

func (c *settingsContent) OnNav(ctx app.Context) {
	var enUser string
	var user models.User

	ctx.LocalStorage().Get("user", &enUser)

	// decode, decrypt and unmarshal the local storage string
	if err := prepare(enUser, &user); err != nil {
		c.toastText = "frontend decoding/decryption failed: " + err.Error()
		c.toastShow = true
		return
	}

	c.user = user
	c.loggedUser = user.Nickname
}

func (c *settingsContent) onClickPass(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)

		if passphrase == "" || passphraseAgain == "" {
			toastText = "both passphrases need to be filled, or text changed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if passphrase != passphraseAgain {
			toastText = "passphrases do not match"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		passHash := sha512.Sum512([]byte(passphrase + config.Pepper))

		if _, ok := litterAPI("PUT", "/api/users", models.User{
			Nickname:   c.user.Nickname,
			Passphrase: string(passHash[:]),
			About:      c.user.About,
			Email:      c.user.Email,
		}); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Navigate("/users")
	})
}

func (c *settingsContent) onClickAbout(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		aboutText := strings.TrimSpace(c.aboutText)

		if aboutText == "" {
			toastText = "about textarea needs to be filled, or you prolly haven't changed the text"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if len(aboutText) > 100 {
			toastText = "about text has to be shorter than 100 chars"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if _, ok := litterAPI("PUT", "/api/users", models.User{
			Nickname:   c.user.Nickname,
			Passphrase: c.user.Passphrase,
			About:      aboutText,
			Email:      c.user.Email,
		}); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		c.user.About = c.aboutText

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toastText = "cannot update local storage"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.LocalStorage().Set("user", userStream)

		ctx.Navigate("/users")
	})
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = (c.toastText != "")
	c.settingsButtonDisabled = false
}

func (c *settingsContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("change your passphrase, or your bottom text"),

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

		// password change
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphrase)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseAgain)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase again").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClickPass).Disabled(c.settingsButtonDisabled),

		app.Div().Class("large-divider"),

		// about textarea
		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Text(c.user.About).Class("active").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout).Disabled(c.settingsButtonDisabled),
		app.Div().Class("space"),
	)
}
