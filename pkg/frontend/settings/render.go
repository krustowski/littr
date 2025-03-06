package settings

import (
	"log"
	"net/url"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/frontend/atomic/organisms"
)

var textboxStrings = map[string]string{
	"avatar-gravatar": "",
}

func (c *Content) Render() app.UI {
	devicesToShow := len(c.devices)

	return app.Main().Class("responsive").ID("anchor-settings-top").Body(
		&atoms.PageHeading{
			Title: "settings",
		},

		//
		// Section user and avatar
		//

		&atoms.PageHeading{
			Title: "user and avatar",
			Level: 6,
		},

		// Logged user's info.
		app.Article().Class("row border thicc").Body(
			app.I().Text("person").Class(""),
			app.If(c.user.Nickname != "", func() app.UI {
				return app.P().Class("max").Body(
					app.Span().Text("Logged as: "),
					app.Span().Class("bold blue-text").Text(c.user.Nickname),
					app.Div().Class("small-space"),
					app.Span().Text("E-mail: "),
					app.Span().Class("bold blue-text").Text(c.user.Email),
				)
			}).Else(func() app.UI {
				return app.Progress().Class("circle blue-border active")
			}),
		),

		// Gravatar linking info.
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
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
			//app.Label().Text("image").Class("active blue-text"),
			//app.I().Text("image"),
		),

		// Infobox about Gravatar image's caching.
		app.Div().Class("space"),
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Text("Note: if you just changed your icon at Gravatar.com, and the thumbnail above shows the old avatar, some intercepting cache probably has the resource cached (e.g. your browser). You may need to wait for some time for the change to propagate through the network."),
			),
		),

		app.Div().Class("space"),

		//
		// Section switches
		//

		&atoms.PageHeading{
			Title: "switches",
			Level: 6,
		},

		// Darkmode infobox.
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("blue-text").Text("The UI mode "),
				app.Span().Text("can be adjusted according to the user's choice. The mode may differ on the other sessions when logged-in on multiple devices."),
			),
		),

		// Darkmode switch.
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("dark/light mode switch"),
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
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("blue-text").Text("The local time mode "),
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
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("blue-text").Text("The live mode "),
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
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("blue-text").Text("Private account "),
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

		&atoms.PageHeading{
			Title: "notifications",
			Level: 6,
		},

		// Notification infoboxes.
		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("blue-text").Text("Reply "),
				app.Span().Text("notifications are fired when someone posts a reply to your post."),
				app.Div().Class("small-space"),
				app.Span().Class("blue-text").Text("Mention "),
				app.Span().Text("notifications are fired when someone mentions you via the at-sign (@) handler in their post (e.g. Hello, @example!)."),
			),
		),

		app.Article().Class("row border blue-border thicc info").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Text("You will be prompted for the notification permission, which is required if you want to subscribe to the notification service. Your device's UUID (unique identification string) will be saved in the database to be used by the notification service. You can delete any subscribed device any time (if listed below)."),
			),
		),

		// Subscription deletion modal.
		app.If(c.deleteSubscriptionModalShow, func() app.UI {
			return app.Dialog().ID("delete-modal").Class("grey10 white-text active thicc").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("subscription deletion"),
				),
				app.Div().Class("space"),

				app.Article().Class("row border amber-border white-text thicc warn").Body(
					app.I().Text("warning").Class("amber-text"),
					app.P().Class("max bold").Body(
						app.Span().Text("Are you sure you want to delete this subscription?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max bold black white-text thicc").OnClick(c.onDismissToast).Body(
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("close"),
							app.Text("Cancel"),
						),
					),
					app.Button().Class("max bold red10 white-text thicc").OnClick(c.onClickDeleteSubscription).Body(
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("delete"),
							app.Text("Delete"),
						),
					),
				),
			)
		}),

		// Reply notification switch.
		app.Div().Class("field middle-align").Body(
			app.Div().Class("row").Body(
				app.Div().Class("max").Body(
					app.Span().Text("reply notification switch"),
				),
				app.Label().Class("switch icon").Body(
					// A nasty workaround to ensure the switch to be updated "correctly".
					app.If(c.subscription.Replies, func() app.UI {
						return app.Label().Class("switch icon").Body(
							app.Input().Type("checkbox").ID("reply-notification-switch").Checked(true).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
							app.Span().Body(
								app.I().Text("notifications"),
							),
						)
					}).Else(func() app.UI {
						return app.Label().Class("switch icon").Body(
							app.Input().Type("checkbox").ID("reply-notification-switch").Checked(false).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
							app.Span().Body(
								app.I().Text("notifications"),
							),
						)
					}),
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
					app.If(c.subscription.Mentions, func() app.UI {
						return app.Label().Class("switch icon").Body(
							app.Input().Type("checkbox").ID("mention-notification-switch").Checked(true).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
							app.Span().Body(
								app.I().Text("notifications"),
							),
						)
					}).Else(func() app.UI {
						return app.Label().Class("switch icon").Body(
							app.Input().Type("checkbox").ID("mention-notification-switch").Checked(false).Disabled(c.settingsButtonDisabled).OnChange(c.onClickNotifSwitch),
							app.Span().Body(
								app.I().Text("notifications"),
							),
						)
					}),
				),
			),
		),

		// Print list of subscribed devices.
		app.If(devicesToShow > 0, func() app.UI {
			return app.Div().Body(
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

						deviceText := "Device"
						if dev.UUID == c.thisDeviceUUID {
							deviceText = "This device"
						}

						// Append the webpush endpoint in the heading.
						u, err := url.Parse(dev.Subscription.Endpoint)
						if err != nil {
							log.Println(err.Error())
							return nil
						}
						deviceText += " (" + u.Host + ")"

						// Compose the component to show (a device's infobox).
						return app.Article().Class("border thicc").Body(
							app.Div().Class("row max").Body(
								app.Div().Class("max").Body(
									app.P().Class("bold").Body(
										app.Text(deviceText),
									),
									app.P().Body(
										app.Text("Subscribed to: "),
										app.Span().Text(dev.Tags).Class("blue-text"),
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
			)
		}),

		//
		// Section passphrase change
		//

		&atoms.PageHeading{
			Title: "passphrase change",
			Level: 6,
		},

		app.Div().Class("field label border blue-text thicc").Body(
			app.Input().ID("passphrase-current").Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseCurrent)).MaxLength(50).Attr("autocomplete", "current-password"),
			app.Label().Text("Old passphrase").Class("active blue-text"),
		),

		app.Div().Class("field label border blue-text thicc").Body(
			app.Input().ID("passphrase-new").Type("password").Class("active").OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Attr("autocomplete", "new-password"),
			app.Label().Text("New passphrase").Class("active blue-text"),
		),

		app.Div().Class("field label border blue-text thicc").Body(
			app.Input().ID("passphrase-new-again").Type("password").Class("active").OnChange(c.ValueTo(&c.passphraseAgain)).MaxLength(50).Attr("autocomplete", "new-password"),
			app.Label().Text("New passphrase again").Class("active blue-text"),
		),

		app.Div().Class("row").Body(
			app.Button().Class("max shrink center primary-container white-text bold thicc").OnClick(c.onClickPass).Disabled(c.settingsButtonDisabled).Body(
				app.Span().Body(
					app.I().Style("padding-right", "5px").Text("save"),
					app.Text("Save"),
				),
			),
		),

		app.Div().Class("space"),

		//
		// Section about-you
		//

		&atoms.PageHeading{
			Title: "about you",
			Level: 6,
		},

		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			Text:        "This textarea is to hold your current status, a brief info about you, or just anything up to 100 characters.",
			MakeSummary: true,
		},

		app.Div().Class("space"),

		&atoms.Textarea{
			ID:               "about-you-textarea",
			Class:            "field textarea label border extra blue-text thicc",
			Content:          c.aboutText,
			LabelText:        "About",
			OnBlurActionName: "blur-about-textarea",
		},

		&atoms.Button{
			Class:             "max responsive shrink primary-container white-text bold thicc",
			OnClickActionName: "about-you-submit",
			Disabled:          c.settingsButtonDisabled,
			Icon:              "save",
			Text:              "Save",
		},

		app.Div().Class("space"),

		//
		// Section website
		//

		&atoms.PageHeading{
			Title: "website link",
			Level: 6,
		},

		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			Text:        "Down below, you can enter a link to your personal homepage. The link will then be visible to others via the user modal on the users (flowers) page.",
			MakeSummary: true,
		},

		app.Div().Class("space"),

		app.Div().Class("field label border blue-text thicc").Body(
			app.Label().Text("URL").Class("active blue-text"),
			&atoms.Input{
				ID:           "website-input",
				Type:         "text",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Value:        c.website,
				AutoComplete: true,
				MaxLength:    60,
			},
			//app.Input().ID("website-input").Type("text").Class("active").OnChange(c.ValueTo(&c.website)).AutoComplete(true).MaxLength(60).Value(c.user.Web),
		),

		&atoms.Button{
			Class:             "max responsive shrink primary-container white-text bold thicc",
			OnClickActionName: "website-input",
			Disabled:          c.settingsButtonDisabled,
			Icon:              "save",
			Text:              "Save",
		},

		app.Div().Class("space"),

		//
		// Section account deletion
		//

		&atoms.PageHeading{
			Title: "account deletion",
			Level: 6,
		},

		&organisms.ModalUserDelete{
			ModalShow:                      c.deleteAccountModalShow,
			LoggedUserNickname:             c.user.Nickname,
			OnClickDismissActionName:       "dismiss",
			OnClickDeleteAccountActionName: "user-delete",
		},

		&molecules.TextBox{
			Class:     "row border red-border thicc danger",
			Icon:      "warning",
			IconClass: "red-text",
			Text:      "Please note that this action is irreversible!",
		},

		app.Div().Class("space"),

		&atoms.Button{
			Class:             "max responsive shrink red10 white-text bold thicc",
			Disabled:          c.settingsButtonDisabled,
			OnClickActionName: "user-modal-delete-show",
			Icon:              "delete",
			Text:              "Delete",
		},

		app.Div().Class("large-space"),
	)
}
