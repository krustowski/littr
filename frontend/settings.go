package frontend

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"log"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type SettingsPage struct {
	app.Compo
}

type settingsContent struct {
	app.Compo

	loggedUser string

	passphrase      string
	passphraseAgain string
	aboutText       string
	user		models.User

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
	ctx.LocalStorage().Get("userName", &c.loggedUser)

	var enUser string
	var preUser []byte
	var decUser []byte
	var user models.User

	ctx.LocalStorage().Get("user", &enUser)
	preUser = config.Decrypt([]byte(config.Pepper), []byte(enUser))

	// beware base64 being used by the framework/browser
	decUser, err := base64.StdEncoding.DecodeString(string(preUser))
	if err != nil {
		c.toastText = "frontend decoding failed: " + err.Error()
		c.toastShow = true
		return
	}
	
	if err := json.Unmarshal(decUser, &user); err != nil {
		c.toastText = "frontend unmarshal failed: " + err.Error()
		c.toastShow = true
		return
	}

	log.Println(user.Nickname + user.About)

	c.user = user
}

func (c *settingsContent) onClickPass(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		c.toastShow = true
		if c.passphrase == "" || c.passphraseAgain == "" {
			c.toastText = "both passphrases need to be filled"
			return
		}

		if c.passphrase != c.passphraseAgain {
			c.toastText = "passphrases do not match"
			return
		}

		passHash := sha512.Sum512([]byte(c.passphrase + config.Pepper))

		if _, ok := litterAPI("PUT", "/api/users", models.User{
			Nickname:   c.user.Nickname,
			Passphrase: string(passHash[:]),
			About:      c.user.About,
			Email:      c.user.Email,
		}); !ok {
			c.toastText = "generic backend error"
			return
		}

		c.toastShow = false
		ctx.Navigate("/users")
	})
}

func (c *settingsContent) onClickAbout(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.aboutText == "" {
			c.toastShow = true
			c.toastText = "about textarea needs to be filled"
			return
		}

		if len(c.aboutText) > 100 {
			c.toastShow = true
			c.toastText = "about text has to be shorter than 100 chars"
			return
		}

		if _, ok := litterAPI("PUT", "/api/users", models.User{
			Nickname:   c.user.Nickname,
			Passphrase: c.user.Passphrase,
			About:      c.aboutText,
			Email:      c.user.Email,
		}); !ok {
			c.toastShow = true
			c.toastText = "generic backend error"
			return
		}

		c.toastShow = false
		ctx.Navigate("/users")
	})
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *settingsContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
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
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphrase)),
			app.Label().Text("passphrase"),
		),
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("password").OnChange(c.ValueTo(&c.passphraseAgain)),
			app.Label().Text("passphrase again"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClickPass),

		app.Div().Class("large-divider"),

		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Text("").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout),
	)
}
