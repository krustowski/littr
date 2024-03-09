package frontend

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/SherClockHolmes/webpush-go"
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
	toastType string

	darkModeOn   bool
	replyNotifOn bool

	notificationPermission app.NotificationPermission
	subscribed             bool

	settingsButtonDisabled bool

	deleteAccountModalShow bool
}

func (p *SettingsPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&settingsContent{},
	)
}

func (p *SettingsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("settings / littr")

	ctx.LocalStorage().Get("mode", &p.mode)
}

func (c *settingsContent) OnMount(ctx app.Context) {
	c.notificationPermission = ctx.Notifications().Permission()
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

		if data, ok := litterAPI("GET", "/api/users", nil, user.Nickname, 0); ok {
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
		//ctx.LocalStorage().Set("mode", user.AppBgMode)

		ctx.Dispatch(func(ctx app.Context) {
			c.user = updatedUser
			c.loggedUser = user.Nickname

			c.darkModeOn = mode == "dark"
			//c.darkModeOn = user.AppBgMode == "dark"

			c.replyNotifOn = c.notificationPermission == app.NotificationGranted
			//c.replyNotifOn = user.ReplyNotificationOn
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

		passHash := sha512.Sum512([]byte(passphrase + app.Getenv("APP_PEPPER")))
		updatedUser := c.user
		//updatedUser.Passphrase = string(passHash[:])
		updatedUser.PassphraseHex = fmt.Sprintf("%x", passHash)

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if data, ok := litterAPI("PUT", "/api/users", updatedUser, c.user.Nickname, 0); !ok {
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

		if _, ok := litterAPI("PUT", "/api/users", updatedUser, c.user.Nickname, 0); !ok {
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

		if _, ok := litterAPI("PUT", "/api/users", updatedUser, c.user.Nickname, 0); !ok {
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

	ctx.Async(func() {
		if _, ok := litterAPI("DELETE", "/api/users", c.user, c.user.Nickname, 0); !ok {
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

func (c *settingsContent) onReplyNotifSwitch(ctx app.Context, e app.Event) {
	checked := ctx.JSSrc().Get("checked").Bool()
	toastText := ""

	//c.replyNotifOn = !c.replyNotifOn

	// unsubscribe
	if !checked {
		uuid := ctx.DeviceID()

		payload := struct {
			UUID string `json:"device_uuid"`
		}{
			UUID: uuid,
		}

		ctx.Async(func() {
			if _, ok := litterAPI("DELETE", "/api/push", payload, c.user.Nickname, 0); !ok {
				toastText := "cannot reach backend!"

				ctx.Dispatch(func(ctx app.Context) {
					//c.toastText = toastText
					c.toastText = "failed to unsubscribe, try again later"
					c.toastShow = toastText != ""

					c.subscribed = true
				})
				return

			}

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "successfully unsubscribed, notifications off"
				c.toastShow = toastText != ""
				c.toastType = "info"

				c.subscribed = false
			})
			return
		})
	}

	ctx.Async(func() {
		// notify user that their browser does not support notifications and therefore they cannot
		// use notifications service
		if c.notificationPermission == app.NotificationNotSupported && checked {
			toastText = "notifications are not supported in this browser"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")

				c.replyNotifOn = false
				c.subscribed = false
			})
			return
		}

		// request the permission on default when switch is toggled
		if (c.notificationPermission == app.NotificationDefault && checked) ||
			(c.notificationPermission == app.NotificationDenied) {
			c.notificationPermission = ctx.Notifications().RequestPermission()
		}

		// fail on denied permission
		if c.notificationPermission != app.NotificationGranted {
			toastText = "notification permission denied by user"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")

				c.replyNotifOn = false
				c.subscribed = false
			})
			return
		}

		// fetch the unique identificator for notifications and unsubscribe option
		uuid := ctx.DeviceID()

		// VAPID key pair
		privKey := c.user.VapidPrivKey
		pubKey := c.user.VapidPubKey

		// generate the VAPID key pair if missing
		// TODO: move this to backend!
		if privKey == "" || pubKey == "" {
			var err error
			privKey, pubKey, err = webpush.GenerateVAPIDKeys()
			if err != nil {
				toastText = "cannot generate new VAPID keys"

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = toastText != ""

					c.subscribed = false
				})
				return
			}
		}

		// register the subscription
		sub, err := ctx.Notifications().Subscribe(pubKey)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = "failed to subscribe to notifications: " + err.Error()
				c.toastShow = c.toastText != ""

				c.subscribed = false
			})
			return
		}

		// compose the Device struct to be saved to the database
		deviceSub := models.Device{
			UUID:         uuid,
			TimeCreated:  time.Now(),
			Subscription: sub,
		}

		user := c.user
		user.VapidPrivKey = privKey
		user.VapidPubKey = pubKey

		// update user on backend via API
		if _, ok := litterAPI("PUT", "/api/users", user, c.user.Nickname, 0); !ok {
			toastText := "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = toastText != ""

				c.subscribed = false
			})
			return
		}

		// send the registeration to backend
		if _, ok := litterAPI("POST", "/api/push", deviceSub, c.user.Nickname, 0); !ok {
			toastText := "cannot reach backend!"

			ctx.Dispatch(func(ctx app.Context) {
				//c.toastText = toastText
				c.toastText = "failed to subscribe to notifications"
				c.toastShow = toastText != ""

				c.subscribed = false
			})
			return
		}

		// dispatch the good news to client
		ctx.Dispatch(func(ctx app.Context) {
			c.user = user
			c.subscribed = true

			c.toastText = "successfully subscribed"
			c.toastShow = toastText != ""
			c.toastType = "success"
		})

		/*ctx.Notifications().New(app.Notification{
			Title: "littr",
			Icon:  "/web/apple-touch-icon.png",
			Body:  "successfully subscribed to notifications",
			Path:  "/flow",
		})*/

		// encode subscription into a HTTP request body
		/*var body bytes.Buffer
		if err := json.NewEncoder(&body).Encode(sub); err != nil {
			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = "encoding notification subscription failed:" + err.Error()
				c.toastShow = c.toastText != ""

				c.subscribed = false
			})
			return
		}*/
	})
}

func (c *settingsContent) onDarkModeSwitch(ctx app.Context, e app.Event) {
	//m := ctx.JSSrc().Get("checked").Bool()

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
				app.Div().Class("snackbar white-text top active "+toastColor).Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		app.Div().Class("large-divider"),
		app.H6().Text("switches"),
		app.Div().Class("space"),

		// darkmode infobox
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("the UI mode "),
				app.Span().Text("can be adjusted according to the user's input (option) --- experimental, the mode may differ on other browsers (when logged-in on multiple devices)"),
			),
		),

		// darkmode switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("light/dark mode switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("dark-mode-switch").Checked(c.darkModeOn).OnChange(c.onDarkModeSwitch).Disabled(c.settingsButtonDisabled),
					app.Span().Body(
						app.I().Text("dark_mode"),
					),
				),
			),
		),

		// left-hand infobox
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("left-hand switch "),
				app.Span().Text("is a theoretical feature which would enable an user to flip the UI for left-handed folks to browse more smoothly"),
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

		// live infobox
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("live mode "),
				app.Span().Text("is a theoretical feature for the live flow preview experience --- one would see other posts incoming as they reach the backend (new posts rendered in live)"),
			),
		),

		// live switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("live switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("live-switch").Checked(false).Disabled(true).OnChange(nil),
					app.Span().Body(
						app.I().Text("stream"),
					),
				),
			),
		),

		// notification infobox
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("reply notifications "),
				app.Span().Text("are fired when someone posts a reply to your post; you will be notified via your browser as this is the so-called web app"),
			),
		),
		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text("reply notifications "),
				app.Span().Text("enabling the notifications will trigger a request for your browser to allow notifications from littr, and will be enabled until you remove the permission in your browser only"),
			),
		),

		// notification switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("reply notification switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("reply-notification-switch").Checked(c.subscribed).Disabled(c.settingsButtonDisabled).OnChange(c.onReplyNotifSwitch),
					app.Span().Body(
						app.I().Text("notifications"),
					),
				),
			),
		),

		// user avatar change
		app.Div().Class("large-divider"),
		app.H6().Text("change user's avatar"),
		app.Div().Class("space"),

		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Text("one's avatar is linked to one's e-mail address, which has to be registered with "),
				app.A().Class("bold").Text("Gravatar.com").Href("https://gravatar.com/profile/avatars"),
			),
		),
		app.Div().Class("").Body(
			app.P().Text("your e-mail address:"),
			app.Article().Body(
				app.Text(c.user.Email),
			),
			app.P().Text("current avatar:"),
		),

		app.Div().Class("transparent middle-align center-align bottom").Body(
			app.Img().Class("small-width middle-align center-align").Src(c.user.AvatarURL).Style("max-width", "120px").Style("border-radius", "50%"),
		),

		app.Article().Class("row border").Body(
			app.I().Text("lightbulb"),
			app.P().Class("max").Body(
				app.Span().Text("note: if you just changed your icon at Gravatar.com, and the thumbnail above shows the old avatar, some intercepting cache probably has the resource cached --- you need to wait for some time for the change to propagate through the network"),
			),
		),

		// password change
		app.Div().Class("large-divider"),
		app.H6().Text("password change"),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphrase)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseAgain)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase again").Class("active deep-orange-text"),
		),

		app.Button().Class("responsive deep-orange7 white-text bold").Text("change passphrase").OnClick(c.onClickPass).Disabled(c.settingsButtonDisabled),

		// about textarea
		app.Div().Class("large-divider"),
		app.H6().Text("about text change"),
		app.Div().Class("space"),

		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Text(c.user.About).Class("active").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about").Class("active deep-orange-text"),
		),

		app.Button().Class("responsive deep-orange7 white-text bold").Text("change about").OnClick(c.onClickAbout).Disabled(c.settingsButtonDisabled),

		// website link
		app.Div().Class("large-divider"),
		app.H6().Text("website link change"),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("text").Class("active").OnChange(c.ValueTo(&c.website)).AutoComplete(true).MaxLength(60).Value(c.user.Web),
			app.Label().Text("website URL").Class("active deep-orange-text"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("change website").OnClick(c.onClickWebsite).Disabled(c.settingsButtonDisabled),

		// user deletion
		app.Div().Class("large-divider"),
		app.H6().Text("account deletion"),

		app.Article().Class("row border").Body(
			app.I().Text("warning"),
			app.P().Class("max").Text("down here, you can delete your account; please note that this action is irreversible!"),
		),
		app.Div().Class("space"),

		app.Button().Class("responsive red9 white-text bold").Text("delete account").OnClick(c.onClickDeleteAccountModalShow).Disabled(c.settingsButtonDisabled),

		app.Div().Class("large-space"),
	)
}
