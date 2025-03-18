package reset

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.If(!c.showUUIDPage, func() app.UI {
					return app.H5().Text("reset request")
				}).Else(func() app.UI {
					return app.H5().Text("reset confirmation")
				}),
			),
		),

		app.Div().Class("space"),

		//
		// Passphrase request --- insert an e-mail
		//
		app.If(!c.showUUIDPage, func() app.UI {
			// Passphrase reset request lightbulb.
			return app.Div().Body(
				app.Article().Class("row border blue-border info thicc").Body(
					app.I().Text("info").Class("blue-text"),
					app.P().Class("max").Body(
						app.Span().Text("To request a passphrase change, enter your registration e-mail address below, which is linked with your account. A confirmation mail will then be sent into your inbox."),
					),
				),
				app.Div().Class("space"),

				// Passphrase reset credentials fields.
				app.Div().Class("field border label primary-text thicc").Body(
					app.Input().ID("email-input").Type("email").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.email)).Class("active").Attr("autocomplete", "email").AutoFocus(true),
					app.Label().Text("E-mail").Class("active primary-text"),
				),

				// Request button.
				app.Div().Class("row center-align max").Body(
					app.Button().ID("request-button").Class("max shrink primary-container bold thicc").OnClick(c.onClickRequest).Disabled(c.buttonsDisabled).TabIndex(2).Body(
						app.If(c.buttonsDisabled, func() app.UI {
							return app.Progress().Class("circle white-border small")
						}),
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("password"),
							app.Text("Request"),
						),
					),
				),
			)
		},

		//
		// Passphrase reset --- insert the UUID
		//
		).Else(func() app.UI {
			return app.Div().Body(
				// Passphrase reset lightbulb.
				app.Article().Class("row border blue-border info thicc").Body(
					app.I().Text("info").Class("blue-text"),
					app.P().Class("max").Body(
						app.Span().Text("Enter the UUID code which has been sent to your inbox. If the code is correct, your passphrase will be automatically regenerated and another confirmation mail containing the passphrase will be sent to your e-mail address."),
					),
				),
				app.Div().Class("space"),

				// Passphrase reset credentials fields.
				app.Div().Class("field border label primary-text thicc").Body(
					app.Input().ID("uuid-input").Type("text").Required(true).TabIndex(1).Value("").OnChange(c.ValueTo(&c.uuid)).Class("active").AutoFocus(true),
					app.Label().Text("UUID").Class("active primary-text"),
				),

				// Passphrase reset button.
				app.Div().Class("row center-align max").Body(
					app.Button().ID("reset-button").Class("max shrink primary-container primary-text bold thicc").TabIndex(2).OnClick(c.onClickReset).Disabled(c.buttonsDisabled).Body(
						app.If(c.buttonsDisabled, func() app.UI {
							return app.Progress().Class("circle white-border small")
						}),
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("password"),
							app.Text("Reset"),
						),
					),
				),
			)
		},
		),

		app.Div().Class("space"),
	)
}
