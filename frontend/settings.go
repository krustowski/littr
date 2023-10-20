package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"log"
	"net/url"
	"regexp"
	"strings"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type SettingsPage struct {
	app.Compo

	mode string
}

type settingsContent struct {
	app.Compo

	// TODO: review this
	loggedUser string

	// used with forms
	passphrase      string
	passphraseAgain string
	aboutText       string
	website         string

	// loaded logged user's struct
	user models.User

	// message toast vars
	toastShow bool
	toastText string

	darkModeOn bool

	settingsButtonDisabled bool

	deleteAccountModalShow bool
}

func (p *SettingsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class(p.mode),
		&header{},
		&footer{},
		&settingsContent{},
	)
}

func (p *SettingsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("settings / littr")
	ctx.LocalStorage().Get("mode", &p.mode)
}

func (c *settingsContent) OnNav(ctx app.Context) {
	toastText := ""

	var enUser string
	var user models.User

	ctx.Async(func() {
		ctx.LocalStorage().Get("user", &enUser)

		// decode, decrypt and unmarshal the local storage string
		if err := prepare(enUser, &user); err != nil {
			toastText = "frontend decoding/decryption failed: " + err.Error()
			return
		}

		usersPre := struct {
			Users map[string]models.User `json:"users"`
		}{}

		if data, ok := litterAPI("GET", "/api/users", nil, user.Nickname); ok {
			err := json.Unmarshal(*data, &usersPre)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parse error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			toastText = "cannot fetch users list"
			log.Println(toastText)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		updatedUser := usersPre.Users[user.Nickname]

		// get the mode
		var mode string
		ctx.LocalStorage().Get("mode", &mode)

		ctx.Dispatch(func(ctx app.Context) {
			c.user = updatedUser
			c.loggedUser = user.Nickname
			c.darkModeOn = mode == "dark"

			//c.darkModeOn = user.AppBgMode == "dark"
		})
		return
	})

	return
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
		updatedUser := c.user
		updatedUser.Passphrase = string(passHash[:])

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if data, ok := litterAPI("PUT", "/api/users", updatedUser, c.user.Nickname); !ok {
			if err := json.Unmarshal(*data, &response); err != nil {
				toastText = "JSON parse error: " + err.Error()
			}
			toastText = "generic backend error: " + response.Message

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		c.user.Passphrase = string(passHash[:])

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toastText = "cannot update local storage"

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

		updatedUser := c.user
		updatedUser.About = aboutText

		if _, ok := litterAPI("PUT", "/api/users", updatedUser, c.user.Nickname); !ok {
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

func (c *settingsContent) onClickWebsite(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	ctx.Async(func() {
		website := strings.TrimSpace(c.website)

		// check the trimmed version of website string
		if website == "" {
			toastText = "website URL has to be filled, or changed"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// check the URL/URI format
		if _, err := url.ParseRequestURI(website); err != nil {
			toastText = "website prolly not a valid URL"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// create a regex object
		regex, err := regexp.Compile("^(http|https)://")
		if err != nil {
			toastText := "failed to check the website (regex object fail)"
			log.Println(toastText)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if !regex.MatchString(website) {
			website = "https://" + website
		}

		updatedUser := c.user
		updatedUser.Web = website

		if _, ok := litterAPI("PUT", "/api/users", updatedUser, c.user.Nickname); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		c.user.Web = c.website

		var userStream []byte
		if err := reload(c.user, &userStream); err != nil {
			toastText = "cannot update local storage"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Navigate("/users")
	})
	return
}

func (c *settingsContent) onClickDeleteAccount(ctx app.Context, e app.Event) {
	toastText := ""

	c.settingsButtonDisabled = true

	//ctx.LocalStorage().Set("userLogged", false)
	ctx.LocalStorage().Set("user", "")
	//h.userLogged = false

	ctx.Async(func() {
		if _, ok := litterAPI("DELETE", "/api/users", c.user, c.user.Nickname); !ok {
			toastText = "generic backend error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		//c.userLogged = false
		//ctx.Navigate("/login")
		ctx.Navigate("/logout")
	})
	return
}

func (c *settingsContent) onDarkModeSwitch(ctx app.Context, e app.Event) {

	m := ctx.JSSrc().Get("value").String()
	log.Println(m)

	c.darkModeOn = !c.darkModeOn

	ctx.LocalStorage().Set("mode", "dark")
	if !c.darkModeOn {
		ctx.LocalStorage().Set("mode", "light")
	}

	//c.app.Window().Get("body").Call("toggleClass", "lightmode")
}

func (c *settingsContent) onClickDeleteAccountModalShow(ctx app.Context, e app.Event) {
	c.deleteAccountModalShow = true
	c.settingsButtonDisabled = true
}

func (c *settingsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastText = ""
	c.toastShow = (c.toastText != "")
	c.settingsButtonDisabled = false
	c.deleteAccountModalShow = false
}

func (c *settingsContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("littr settings").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("change your passphrase, or your bottom text"),

		app.Div().Class("space"),

		// acc deletion modal
		app.If(c.deleteAccountModalShow,
			app.Dialog().Class("grey9 white-text active").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("account deletion"),
				),
				app.P().Text("are you sure you want to delete your account and all posted items?"),
				app.Div().Class("space"),
				app.Nav().Class("center-align").Body(
					app.Button().Class("border deep-orange7 white-text").Text("yeah").OnClick(c.onClickDeleteAccount),
					app.Button().Class("border deep-orange7 white-text").Text("nope").OnClick(c.dismissToast),
				),
			),
		),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		app.Div().Class("large-divider"),
		app.H5().Text("switches"),
		app.Div().Class("space"),

		// darkmode switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("light/dark mode switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("dark-mode-switch").Checked(c.darkModeOn).OnChange(c.onDarkModeSwitch),
					app.Span().Body(
						app.I().Text("dark_mode"),
					),
				),
			),
		),

		// left-hand switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("left-hand switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("left-hand-switch").Checked(false).Disabled(true).OnChange(nil),
					app.Span().Body(
						app.I().Text("front_hand"),
					),
				),
			),
		),

		app.Div().Class("large-divider"),
		app.H5().Text("password change"),
		app.Div().Class("space"),

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
		app.H6().Text("about text change"),
		app.Div().Class("space"),

		// about textarea
		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Text(c.user.About).Class("active").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about").Class("active"),
		),

		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout).Disabled(c.settingsButtonDisabled),

		app.Div().Class("large-divider"),
		app.H5().Text("website link change"),
		app.Div().Class("space"),

		// website link
		app.Div().Class("field label border invalid deep-orange-text").Body(
			app.Input().Type("text").Class("active").OnChange(c.ValueTo(&c.website)).AutoComplete(true).MaxLength(60).Value(c.user.Web),
			app.Label().Text("website URL").Class("active"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change website").OnClick(c.onClickWebsite).Disabled(c.settingsButtonDisabled),

		app.Div().Class("large-divider"),
		app.H5().Text("account deletion"),
		app.P().Text("down here, you can delete your account; please note that this action is irreversible!"),
		app.Div().Class("space"),

		app.Button().Class("responsive red9 white-text bold").Text("delete account").OnClick(c.onClickDeleteAccountModalShow).Disabled(c.settingsButtonDisabled),

		app.Div().Class("large-space"),
	)
}
