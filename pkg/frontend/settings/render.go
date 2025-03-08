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

		&molecules.TextBox{
			Class:      "row border thicc",
			Icon:       "person",
			MarkupText: FormatUserInfo,
			FormatArgs: []interface{}{c.user.Nickname, c.user.Email},
			ShowLoader: c.user.Nickname == "",
		},

		app.Div().Class("space"),

		// User's avatar view and a (hidden) upload option.
		app.Div().Class("transparent middle-align center-align bottom").Body(
			app.Img().Class("small-width middle-align center-align").Src(c.user.AvatarURL).Style("max-width", "120px").Style("border-radius", "50%"),
			app.Input().ID("avatar-upload").Class("active").Type("file").OnInput(c.onAvatarChange).Accept("image/png, image/jpeg"),
		),

		// Gravatar linking info.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  InfoGravatarLinking,
			MakeSummary: true,
		},

		app.Div().Class("space"),

		//
		// Section switches
		//

		&atoms.PageHeading{
			Title: "switches",
			Level: 6,
		},

		// Darkmode infobox.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  InfoUIMode,
			MakeSummary: true,
		},

		// UI mode switch.
		&molecules.Switch{
			Icon:               "dark_mode",
			ID:                 "ui-mode-switch",
			Text:               "UI mode switch",
			Checked:            c.user.Options["uiMode"],
			Disabled:           c.settingsButtonDisabled,
			OnChangeActionName: "options-switch-change",
		},

		// Local time infobox.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  InfoLocalTimeMode,
			MakeSummary: true,
		},

		// Local time switch
		&molecules.Switch{
			Icon:               "schedule",
			ID:                 "local-time-mode-switch",
			Text:               "local time mode switch",
			Checked:            c.user.Options["localTimeMode"],
			Disabled:           c.settingsButtonDisabled,
			OnChangeActionName: "options-switch-change",
		},

		// Live mode infobox.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  InfoLiveMode,
			MakeSummary: true,
		},

		// Live mode switch.
		&molecules.Switch{
			Icon:               "stream",
			ID:                 "live-mode-switch",
			Text:               "live mode switch",
			Checked:            c.user.Options["liveMode"],
			Disabled:           c.settingsButtonDisabled,
			OnChangeActionName: "options-switch-change",
		},

		// Private account infobox.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  InfoPrivateMode,
			MakeSummary: true,
		},

		// Private account switch.
		&molecules.Switch{
			Icon:               "lock",
			ID:                 "private-mode-switch",
			Text:               "private mode switch",
			Checked:            c.user.Options["private"],
			Disabled:           c.settingsButtonDisabled,
			OnChangeActionName: "options-switch-change",
		},

		//
		// Section notifications
		//

		&atoms.PageHeading{
			Title: "notifications",
			Level: 6,
		},

		&organisms.ModalSubscriptionDelete{
			ModalShow:                c.deleteSubscriptionModalShow,
			OnClickDismissActionName: "dismiss",
			OnClickDeleteActionName:  "subscription-delete",
		},

		// Notification infoboxes.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  InfoNotifications,
			MakeSummary: true,
		},

		// Reply notification switch.
		&molecules.Switch{
			Icon:               "notifications",
			ID:                 "reply-notif-switch",
			Text:               "reply notification switch",
			Checked:            c.subscription.Replies,
			Disabled:           c.settingsButtonDisabled,
			OnChangeActionName: "notifs-switch-change",
		},

		// Mention notification switch.
		&molecules.Switch{
			Icon:               "notifications",
			ID:                 "mention-notif-switch",
			Text:               "mention notification switch",
			Checked:            c.subscription.Mentions,
			Disabled:           c.settingsButtonDisabled,
			OnChangeActionName: "notifs-switch-change",
		},

		// Print list of subscribed devices.
		app.If(devicesToShow > 0, func() app.UI {
			return app.Div().Body(
				&atoms.PageHeading{
					Title: "subscribed devices",
					Level: 6,
				},

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
						return &molecules.TextBox{
							Class:      "row border deep-orange-border thicc",
							Icon:       "",
							IconClass:  "deep-orange-text",
							MarkupText: InfoSubscribedDevice,
							FormatArgs: []interface{}{deviceText, dev.Tags, dev.TimeCreated.Format("2006-01-02 15:04:05")},
							Button: &atoms.Button{
								ID:                dev.UUID,
								Class:             "transparent circle",
								OnClickActionName: "subscription-delete-modal-show",
								Disabled:          c.settingsButtonDisabled,
								Icon:              "delete",
							},
						}

						/*return app.Article().Class("border thicc").Body(
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
						)*/
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
			&atoms.Input{
				ID:           "passphrase-current",
				Type:         "password",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.passphraseCurrent,
				AutoComplete: false,
				MaxLength:    64,
				Attr:         map[string]string{"autocomplete": "current-password"},
			},
			app.Label().Text("Old passphrase").Class("active blue-text"),
		),

		app.Div().Class("field label border blue-text thicc").Body(
			&atoms.Input{
				ID:           "passphrase-new",
				Type:         "password",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.passphrase,
				AutoComplete: false,
				MaxLength:    64,
				Attr:         map[string]string{"autocomplete": "new-password"},
			},
			app.Label().Text("New passphrase").Class("active blue-text"),
		),

		app.Div().Class("field label border blue-text thicc").Body(
			&atoms.Input{
				ID:           "passphrase-new-again",
				Type:         "password",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.passphraseAgain,
				AutoComplete: false,
				MaxLength:    64,
				Attr:         map[string]string{"autocomplete": "new-password"},
			},
			app.Label().Text("New passphrase again").Class("active blue-text"),
		),

		&atoms.Button{
			Class:             "max responsive shrink center primary-container white-text bold thicc",
			OnClickActionName: "passphrase-submit",
			Disabled:          c.settingsButtonDisabled,
			Icon:              "save",
			Text:              "Save",
		},

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
			Text:        InfoAboutYouTextarea,
			MakeSummary: true,
		},

		app.Div().Class("space"),

		&atoms.Textarea{
			ID:               "about-you-textarea",
			Class:            "field textarea label border extra blue-text thicc",
			Content:          c.user.About,
			ContentPointer:   &c.aboutText,
			LabelText:        "About",
			OnBlurActionName: "blur-about-textarea",
		},

		&atoms.Button{
			Class:             "max responsive shrink center primary-container white-text bold thicc",
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
			Text:        InfoWebsiteLink,
			MakeSummary: true,
		},

		app.Div().Class("space"),

		app.Div().Class("field label border blue-text thicc").Body(
			&atoms.Input{
				ID:           "website-input",
				Type:         "text",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.user.Web,
				AutoComplete: true,
				MaxLength:    60,
			},
			app.Label().Text("URL").Class("active blue-text"),
			//app.Input().ID("website-input").Type("text").Class("active").OnChange(c.ValueTo(&c.website)).AutoComplete(true).MaxLength(60).Value(c.user.Web),
		),

		&atoms.Button{
			Class:             "max responsive shrink center primary-container white-text bold thicc",
			OnClickActionName: "website-submit",
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
			Text:      AlertUserDeletion,
		},

		app.Div().Class("space"),

		&atoms.Button{
			Class:             "max responsive shrink center red10 white-text bold thicc",
			Disabled:          c.settingsButtonDisabled,
			OnClickActionName: "user-delete-modal-show",
			Icon:              "delete",
			Text:              "Delete",
		},

		app.Div().Class("large-space"),
	)
}
