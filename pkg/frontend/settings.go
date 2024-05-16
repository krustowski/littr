package frontend

import (
	"encoding/json"
	"log"
	"net/url"

	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

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
	subscription           struct {
		Replies  bool
		Mentions bool
	}
	mentionNotificationEnabled bool

	settingsButtonDisabled bool

	deleteAccountModalShow      bool
	deleteSubscriptionModalShow bool

	VAPIDpublic string
	devices     []models.Device
	thisDevice  models.Device

	thisDeviceUUID string
	interactedUUID string

	newFigLink string
	newFigData []byte
	newFigFile string
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
	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	ctx.Async(func() {
		payload := struct {
			Key        string                 `json:"key"`
			PublicKey  string                 `json:"public_key"`
			Users      map[string]models.User `json:"users"`
			Subscribed bool                   `json:"subscribed"`
			Devices    []models.Device        `json:"devices"`
			Code       int                    `json:"code"`
		}{}

		if data, ok := litterAPI("GET", "/api/v1/users", nil, ctx.DeviceID(), 0); ok {
			err := json.Unmarshal(*data, &payload)
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

		if payload.Code == 401 {
			toastText = "please log-in again"

			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		user := payload.Users[payload.Key]

		var thisDevice models.Device
		for _, dev := range payload.Devices {
			if dev.UUID == ctx.DeviceID() {
				thisDevice = dev
				break
			}
		}

		subscription := c.subscription
		if helpers.Contains(thisDevice.Tags, "reply") {
			subscription.Replies = true
		}
		if helpers.Contains(thisDevice.Tags, "mention") {
			subscription.Mentions = true
		}

		// get the mode
		var mode string
		ctx.LocalStorage().Get("mode", &mode)
		//ctx.LocalStorage().Set("mode", user.AppBgMode)

		ctx.Dispatch(func(ctx app.Context) {
			c.user = user
			c.loggedUser = user.Nickname
			c.devices = payload.Devices

			c.subscribed = payload.Subscribed
			c.subscription = subscription

			c.darkModeOn = mode == "dark"
			//c.darkModeOn = user.AppBgMode == "dark"

			c.VAPIDpublic = payload.PublicKey
			c.thisDeviceUUID = ctx.DeviceID()
			c.thisDevice = thisDevice

			c.replyNotifOn = c.notificationPermission == app.NotificationGranted
			//c.replyNotifOn = user.ReplyNotificationOn

			c.settingsButtonDisabled = false
		})
		return
	})
	return
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

	devicesToShow := len(c.devices)

	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("settings"),
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

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("user and avatar"),
			),
		),

		// logged user info
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("person").Class("amber-text"),
			app.If(c.user.Nickname != "",
				app.P().Class("max").Body(
					app.Span().Text("logged as: "),
					app.Span().Class("deep-orange-text").Text(c.user.Nickname),
					app.Div().Class("small-space"),
					app.Span().Text("e-mail: "),
					app.Span().Class("deep-orange-text").Text(c.user.Email),
				),
			).Else(
				app.Progress().Class("circle deep-orange-border active"),
			),
		),

		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Text("one's avatar is linked to one's e-mail address, which has to be registered with "),
				app.A().Class("bold").Text("Gravatar.com").Href("https://gravatar.com/profile/avatars"),
			),
		),
		app.Div().Class("space"),

		// load current user's avatar
		app.Div().Class("transparent middle-align center-align bottom").Body(
			app.Img().Class("small-width middle-align center-align").Src(c.user.AvatarURL).Style("max-width", "120px").Style("border-radius", "50%").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload),
			app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload),
			//app.Input().Class("active").Type("text").Value(c.newFigFile).Disabled(true),
			//app.Label().Text("image").Class("active deep-orange-text"),
			//app.I().Text("image"),
		),

		// infobox about image caching
		app.Div().Class("space"),
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("info").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Text("note: if you just changed your icon at Gravatar.com, and the thumbnail above shows the old avatar, some intercepting cache probably has the resource cached --- you need to wait for some time for the change to propagate through the network"),
			),
		),

		app.Div().Class("space"),
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("switches"),
			),
		),

		// darkmode infobox
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
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

		// live infobox
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
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
					app.Input().Type("checkbox").ID("live-switch").Checked(true).Disabled(true).OnChange(nil),
					app.Span().Body(
						app.I().Text("stream"),
					),
				),
			),
		),

		// private acc infobox
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("private account "),
				app.Span().Text("is a feature allowing one to be hidden on littr in terms of free reachibility via the users page, thus others have to be allowed to add you in their flow and to see your profile"),
			),
		),

		// private acc switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("private acc switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("private-acc-switch").Checked(c.user.Private).Disabled(c.settingsButtonDisabled).OnChange(c.onClickPrivateSwitch),
					app.Span().Body(
						app.I().Text("lock"),
					),
				),
			),
		),

		// notifications
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("notifications"),
			),
		),

		// notification infobox
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("reply notifications "),
				app.Span().Text("are fired when someone posts a reply to your post; you will be notified via your browser as this is the so-called web app"),
				app.Div().Class("small-space"),
				app.Span().Class("deep-orange-text").Text("mention notifications "),
				app.Span().Text("are fired when someone mentions you via the at-sign (@) handler (e.g. @example)"),
			),
		),
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text("reply notifications "),
				//app.Span().Text("enabling the notifications will trigger a request for your browser to allow notifications from littr, and will be enabled until you remove the permission in your browser only"),
				app.Span().Text("by switching this one you will be prompted for the notification permission, which is required to be positive if one wants to subscribe to notifications; this device's UUID will be used to identify this very blackbox --- to route notifications correctly to you"),
			),
		),

		// subs deletion modal
		app.If(c.deleteSubscriptionModalShow,
			app.Dialog().Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("subscription deletion"),
				),
				app.Div().Class("space"),

				app.Article().Class("row surface-container-highest").Body(
					app.I().Text("warning").Class("amber-text"),
					app.P().Class("max").Body(
						app.Span().Text("are you sure you want to delete this subscription?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max border red9 white-text").Text("yeah").Style("border-radius", "8px").OnClick(c.onClickDeleteSubscription),
					app.Button().Class("max border deep-orange7 white-text").Text("nope").Style("border-radius", "8px").OnClick(c.dismissToast),
				),
			),
		),

		// reply notification switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("reply notification switch"),
				),
				app.Label().Class("switch icon").Body(
					// nasty workaround to ensure the switch to be updated "correctly"
					app.If(c.subscription.Replies,
						app.Input().Type("checkbox").ID("reply-notification-switch").Checked(true).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
						app.Span().Body(
							app.I().Text("notifications"),
						),
					).Else(
						app.Input().Type("checkbox").ID("reply-notification-switch").Checked(false).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
						app.Span().Body(
							app.I().Text("notifications"),
						),
					),
				),
			),
		),

		// mention notification switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("mention notification switch"),
				),
				app.Label().Class("switch icon").Body(
					// nasty workaround to ensure the switch to be updated "correctly"
					app.If(c.subscription.Mentions,
						app.Input().Type("checkbox").ID("mention-notification-switch").Checked(true).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
						app.Span().Body(
							app.I().Text("notifications"),
						),
					).Else(
						app.Input().Type("checkbox").ID("mention-notification-switch").Checked(false).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
						app.Span().Body(
							app.I().Text("notifications"),
						),
					),
				),
			),
		),

		// print list of subscribed devices
		app.If(devicesToShow > 0,
			// user avatar change
			//app.Div().Class("large-divider"),
			//app.Div().Class("space"),
			app.Div().Class("row").Body(
				app.Div().Class("max padding").Body(
					app.H6().Text("registered devices"),
				),
			),

			app.Div().Class().Body(
				app.Range(c.devices).Slice(func(i int) app.UI {

					dev := c.devices[i]
					if dev.UUID == "" {
						return nil
					}

					deviceText := "device"
					if dev.UUID == c.thisDeviceUUID {
						deviceText = "this device"
					}

					u, err := url.Parse(dev.Subscription.Endpoint)
					if err != nil {
						log.Println(err.Error())
						return nil
					}
					deviceText += " (" + u.Host + ")"

					return app.Article().Class("surface-container-highest").Style("border-radius", "8px").Body(
						app.Div().Class("row max").Body(
							app.Div().Class("max").Body(

								app.P().Class("bold").Body(app.Text(deviceText)),
								app.P().Body(
									app.Text("subscribed to notifs: "),
									app.Span().Text(dev.Tags).Class("deep-orange-text"),
								),
								app.P().Body(app.Text(dev.TimeCreated)),
							),

							app.Button().ID(dev.UUID).Class("transparent circle").OnClick(c.onClickDeleteSubscriptionModalShow).Disabled(c.settingsButtonDisabled).Body(
								app.I().Text("delete"),
							),
						),
					)
				}),
			),
			app.Div().Class("space"),
		),

		// passphrase change
		//app.Div().Class("space"),
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("passphrase change"),
			),
		),
		//app.Div().Class("medium-space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphrase)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseAgain)).AutoComplete(true).MaxLength(50),
			app.Label().Text("passphrase again").Class("active deep-orange-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max deep-orange7 white-text bold").Text("change passphrase").Style("border-radius", "8px").OnClick(c.onClickPass).Disabled(c.settingsButtonDisabled),
		),

		// about-you textarea
		app.Div().Class("space"),
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("about-you text"),
			),
		),
		//app.Div().Class("medium-space"),

		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Text("this textarea is to hold your status, a brief info about you, just anything up to 100 characters"),
		),
		app.Div().Class("space"),

		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Text(c.user.About).Class("active").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about-you").Class("active deep-orange-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max deep-orange7 white-text bold").Text("change about").Style("border-radius", "8px").OnClick(c.onClickAbout).Disabled(c.settingsButtonDisabled),
		),

		// website link
		app.Div().Class("space"),
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("website link"),
			),
		),
		//app.Div().Class("medium-space"),

		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Text("down below, you can enter a link to your personal homepage --- the link will then be visible to others via the user modal on the users (flowers) page"),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().Type("text").Class("active").OnChange(c.ValueTo(&c.website)).AutoComplete(true).MaxLength(60).Value(c.user.Web),
			app.Label().Text("website URL").Class("active deep-orange-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max deep-orange7 white-text bold").Text("change website").Style("border-radius", "8px").OnClick(c.onClickWebsite).Disabled(c.settingsButtonDisabled),
		),

		// acc deletion modal
		app.If(c.deleteAccountModalShow,
			app.Dialog().Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("account deletion"),
				),
				app.Div().Class("space"),

				app.Article().Class("row surface-container-highest").Body(
					app.I().Text("warning").Class("red-text"),
					app.P().Class("max").Body(
						app.Span().Text("are you sure you want to delete your account and all posted items?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max border red9 white-text").Text("yeah").Style("border-radius", "8px").OnClick(c.onClickDeleteAccount),
					app.Button().Class("max border deep-orange7 white-text").Text("nope").Style("border-radius", "8px").OnClick(c.dismissToast),
				),
			),
		),

		// user deletion
		app.Div().Class("space"),
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("account deletion"),
			),
		),
		//app.Div().Class("space"),

		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("warning").Class("red-text"),
			app.P().Class("max").Text("down here, you can delete your account; please note that this action is irreversible!"),
		),
		app.Div().Class("space"),

		app.Div().Class("row").Body(
			app.Button().Class("max red9 white-text bold").Text("delete account").Style("border-radius", "8px").OnClick(c.onClickDeleteAccountModalShow).Disabled(c.settingsButtonDisabled),
		),

		app.Div().Class("large-space"),
	)
}
