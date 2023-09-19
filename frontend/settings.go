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
	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		passphrase := strings.TrimSpace(c.passphrase)
		passphraseAgain := strings.TrimSpace(c.passphraseAgain)

		if passphrase == "" || passphraseAgain == "" {
			c.toastText = "both passphrases need to be filled, or text changed"
			return
		}

		if passphrase != passphraseAgain {
			c.toastText = "passphrases do not match"
			return
		}

		passHash := sha512.Sum512([]byte(passphrase + config.Pepper))

		if _, ok := litterAPI("PUT", "/api/users", models.User{
			Nickname:   c.user.Nickname,
			Passphrase: string(passHash[:]),
			About:      c.user.About,
			Email:      c.user.Email,
		}); !ok {
			c.toastText = "generic backend error"
			return
		}

		ctx.Navigate("/users")
	})
}

func (c *settingsContent) onClickAbout(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		// trim the padding spaces on the extremities
		// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
		aboutText := strings.TrimSpace(c.aboutText)

		if aboutText == "" {
			c.toastText = "about textarea needs to be filled, or you prolly haven't changed the text"
			return
		}

		if len(aboutText) > 100 {
			c.toastText = "about text has to be shorter than 100 chars"
			return
		}

		if _, ok := litterAPI("PUT", "/api/users", models.User{
			Nickname:   c.user.Nickname,
			Passphrase: c.user.Passphrase,
			About:      aboutText,
			Email:      c.user.Email,
		}); !ok {
			c.toastText = "generic backend error"
			return
		}

		c.user.About = c.aboutText

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			c.toastText = "cannot update local storage"
			return
		}

		ctx.LocalStorage().Set("user", userStream)

		ctx.Navigate("/users")
	})
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = false
}

func (c *settingsContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastText != "" {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("change your passphrase, or your bottom text"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphrase)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase").Class("active"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseAgain)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase again").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClickPass),

		app.Div().Class("large-divider"),

		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Text(c.user.About).Class("active").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout),
	)
}
