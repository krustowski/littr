package settings

import (
	"log"
	"net/url"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	devicesToShow := len(c.devices)

	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("settings"),
			),
		),

		// Snackbar
		app.A().Href(c.toast.TLink).OnClick(c.onDismissToast).Body(
			app.If(c.toast.TText != "",
				app.Div().ID("snackbar").Class("snackbar white-text top active "+common.ToastColor(c.toast.TType)).Style("border-radius", "8px").Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				),
			),
		),

		//
		// Section user and avatar
		//

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("user and avatar"),
			),
		),

		// Logged user's info.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("person").Class("deep-orange-text"),
			app.If(c.user.Nickname != "",
				app.P().Class("max").Body(
					app.Span().Text("Logged as: "),
					app.Span().Class("bold deep-orange-text").Text(c.user.Nickname),
					app.Div().Class("small-space"),
					app.Span().Text("E-mail: "),
					app.Span().Class("bold deep-orange-text").Text(c.user.Email),
				),
			).Else(
				app.Progress().Class("circle deep-orange-border active"),
			),
		),

		// Gravatar linking info.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Text("One's avatar is linked to one's e-mail address, which should be registered with "),
				app.A().Class("bold").Text("Gravatar.com").Href("https://gravatar.com/profile/avatars"),
				app.Span().Text("."),
			),
		),
		app.Div().Class("space"),

		// User's avatar view and a (hidden) upload option.
		app.Div().Class("transparent middle-align center-align bottom").Body(
			app.Img().Class("small-width middle-align center-align").Src(c.user.AvatarURL).Style("max-width", "120px").Style("border-radius", "50%").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload),
			app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload).Accept("image/png, image/jpeg"),
			//app.Input().Class("active").Type("text").Value(c.newFigFile).Disabled(true),
			//app.Label().Text("image").Class("active deep-orange-text"),
			//app.I().Text("image"),
		),

		// Infobox about Gravatar image's caching.
		app.Div().Class("space"),
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Text("Note: if you just changed your icon at Gravatar.com, and the thumbnail above shows the old avatar, some intercepting cache probably has the resource cached (e.g. your browser). You may need to wait for some time for the change to propagate through the network."),
			),
		),

		app.Div().Class("space"),

		//
		// Section switches
		//

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("switches"),
			),
		),

		// Darkmode infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("The UI mode "),
				app.Span().Text("can be adjusted according to the user's choice. The mode may differ on the other sessions when logged-in on multiple devices."),
			),
		),

		// Darkmode switch.
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("Dark/light mode switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("dark-mode-switch").Checked(c.darkModeOn).OnChange(c.onDarkModeSwitch).Disabled(c.settingsButtonDisabled),
					app.Span().Body(
						app.I().Text("dark_mode"),
					),
				),
			),
		),

		// Local time infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("The local time mode "),
				app.Span().Text("is a feature allowing you to see any post's (or poll's) timestamp according to your device's setting (mainly the timezone). When disabled, the server time is used instead."),
			),
		),

		// Local time switch
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("local time mode switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("local-time-mode-switch").Checked(!c.user.LocalTimeMode).OnChange(c.onLocalTimeModeSwitch).Disabled(c.settingsButtonDisabled),
					app.Span().Body(
						app.I().Text("schedule"),
					),
				),
			),
		),

		// Live mode infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("The live mode "),
				app.Span().Text("is a feature for the live flow experience. When enabled, a notice about some followed account's/user's new post is shown on the bottom of the page."),
			),
		),

		// Live mode switch.
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

		// Private account infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("Private account "),
				app.Span().Text("is a feature allowing one to be hidden on the site. When enabled, other accounts/users need to ask you to follow you (the follow request will show on the users page). Any reply to your post will be shown as redacted (a private content notice) to those not following you."),
			),
		),

		// Private account switch.
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("private account switch"),
				),
				app.Label().Class("switch icon").Body(
					app.Input().Type("checkbox").ID("private-acc-switch").Checked(c.user.Private).Disabled(c.settingsButtonDisabled).OnChange(c.onClickPrivateSwitch),
					app.Span().Body(
						app.I().Text("lock"),
					),
				),
			),
		),

		//
		// Section notifications
		//

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("notifications"),
			),
		),

		// Notification infoboxes.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("Reply "),
				app.Span().Text("notifications are fired when someone posts a reply to your post."),
				app.Div().Class("small-space"),
				app.Span().Class("deep-orange-text").Text("Mention "),
				app.Span().Text("notifications are fired when someone mentions you via the at-sign (@) handler in their post (e.g. Hello, @example!)."),
			),
		),

		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Text("You will be prompted for the notification permission, which is required if you want to subscribe to the notification service. Your device's UUID (unique identification string) will be saved in the database to be used by the notification service. You can delete any subscribed device any time (if listed below)."),
			),
		),

		// Subscription deletion modal.
		app.If(c.deleteSubscriptionModalShow,
			app.Dialog().ID("delete-modal").Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("subscription deletion"),
				),
				app.Div().Class("space"),

				app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
					app.I().Text("warning").Class("amber-text"),
					app.P().Class("max").Body(
						app.Span().Text("Are you sure you want to delete this subscription?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max border red10 white-text").Text("yeah").Style("border-radius", "8px").OnClick(c.onClickDeleteSubscription),
					app.Button().Class("max border black white-text").Text("nope").Style("border-radius", "8px").OnClick(c.onDismissToast),
				),
			),
		),

		// Reply notification switch.
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("reply notification switch"),
				),
				app.Label().Class("switch icon").Body(
					// A nasty workaround to ensure the switch to be updated "correctly".
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

		// Mention notification switch.
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("mention notification switch"),
				),
				app.Label().Class("switch icon").Body(
					// A nasty workaround to ensure the switch to be updated "correctly".
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

		// Print list of subscribed devices.
		app.If(devicesToShow > 0,
			app.Div().Class("row").Body(
				app.Div().Class("max padding").Body(
					app.H6().Text("registered devices"),
				),
			),

			// Loop over the array of subscribed devices.
			app.Div().Class().Body(
				app.Range(c.devices).Slice(func(i int) app.UI {
					// Take the i-th device.
					dev := c.devices[i]
					if dev.UUID == "" {
						return nil
					}

					deviceText := "device"
					if dev.UUID == c.thisDeviceUUID {
						deviceText = "this device"
					}

					// Append the webpush endpoint in the heading.
					u, err := url.Parse(dev.Subscription.Endpoint)
					if err != nil {
						log.Println(err.Error())
						return nil
					}
					deviceText += " (" + u.Host + ")"

					// Compose the component to show (a device's infobox).
					return app.Article().Class("surface-container-highest").Style("border-radius", "8px").Body(
						app.Div().Class("row max").Body(
							app.Div().Class("max").Body(
								app.P().Class("bold").Body(
									app.Text(deviceText),
								),
								app.P().Body(
									app.Text("Subscribed to: "),
									app.Span().Text(dev.Tags).Class("deep-orange-text"),
								),
								app.P().Body(
									app.Text("Registered: "),
									app.Text(dev.TimeCreated),
								),
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

		//
		// Section passphrase change
		//

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("passphrase change"),
			),
		),
		//app.Div().Class("medium-space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-current").Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseCurrent)).MaxLength(50).Attr("autocomplete", "current-password"),
			app.Label().Text("old passphrase").Class("active deep-orange-text"),
		),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-new").Type("password").Class("active").OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Attr("autocomplete", "new-password"),
			app.Label().Text("new passphrase").Class("active deep-orange-text"),
		),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-new-again").Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseAgain)).MaxLength(50).Attr("autocomplete", "new-password"),
			app.Label().Text("new passphrase again").Class("active deep-orange-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max shrink center deep-orange7 white-text bold").Text("change passphrase").Style("border-radius", "8px").OnClick(c.onClickPass).Disabled(c.settingsButtonDisabled),
		),

		app.Div().Class("space"),

		//
		// Section about-you
		//

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("about you"),
			),
		),

		// About-you infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Text("This textarea is to hold your current status, a brief info about you, or just anything up to 100 characters."),
		),

		app.Div().Class("space"),

		// About-you textarea
		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().ID("about-you-textarea").Text(c.user.About).Class("active").OnChange(c.ValueTo(&c.aboutText)),
			app.Label().Text("about you").Class("active deep-orange-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max shrink center deep-orange7 white-text bold").Text("change about").Style("border-radius", "8px").OnClick(c.onClickAbout).Disabled(c.settingsButtonDisabled),
		),

		// website link
		app.Div().Class("space"),

		//
		// Section website
		//

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("website link"),
			),
		),

		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Text("Down below, you can enter a link to your personal homepage. The link will then be visible to others via the user modal on the users (flowers) page."),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("website-input").Type("text").Class("active").OnChange(c.ValueTo(&c.website)).AutoComplete(true).MaxLength(60).Value(c.user.Web),
			app.Label().Text("website URL").Class("active deep-orange-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max shrink center deep-orange7 white-text bold").Text("change website").Style("border-radius", "8px").OnClick(c.onClickWebsite).Disabled(c.settingsButtonDisabled),
		),

		app.Div().Class("space"),

		//
		// Section account deletion
		//

		// Account deletion modal.
		app.If(c.deleteAccountModalShow,
			app.Dialog().ID("delete-modal").Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("account deletion"),
				),
				app.Div().Class("space"),

				app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
					app.I().Text("warning").Class("red-text"),
					app.P().Class("max").Body(
						app.Span().Text("Are you sure you want to delete your account and all posted items?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max border red10 white-text").Text("yeah").Style("border-radius", "8px").OnClick(c.onClickDeleteAccount),
					app.Button().Class("max border black white-text").Text("nope").Style("border-radius", "8px").OnClick(c.onDismissToast),
				),
			),
		),

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H6().Text("account deletion"),
			),
		),

		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("warning").Class("red-text"),
			app.P().Class("max").Text("Please note that this action is irreversible!"),
		),
		app.Div().Class("space"),

		app.Div().Class("row").Body(
			app.Button().Class("max shrink center red10 white-text bold").Text("delete account").Style("border-radius", "8px").OnClick(c.onClickDeleteAccountModalShow).Disabled(c.settingsButtonDisabled),
		),

		app.Div().Class("large-space"),
	)
}
