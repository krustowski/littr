package reset

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	toastColor := ""

	switch c.toast.TType {
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
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.If(!c.showUUIDPage,
					app.H5().Text("littr passphrase request"),
				).Else(
					app.H5().Text("littr passphrase reset"),
				),
			),
		),

		app.Div().Class("space"),

		// snackbar
		app.A().Href(c.toast.TLink).OnClick(c.onDismissToast).Body(
			app.If(c.toast.TText != "",
				app.Div().ID("snackbar").Class("snackbar "+toastColor+" white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				),
			),
		),

		// passphrase request --- insert an e-mail
		app.If(!c.showUUIDPage,

			// pwd reset lightbulb
			app.Article().Class("row surface-container-highest").Body(
				app.I().Text("lightbulb").Class("amber-text"),
				app.P().Class("max").Body(
					//app.Span().Class("deep-orange-text").Text(" "),
					app.Span().Text("to request a passphrase change, enter your registration e-mail address below, which is linked with your account; a confirmation mail will then be sent to your inbox"),
				),
			),
			app.Div().Class("space"),

			// pwd reset credentials fields
			app.Div().Class("field border label deep-orange-text").Body(
				app.Input().ID("email-input").Type("email").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.email)).Class("active").Attr("autocomplete", "email").AutoFocus(true),
				app.Label().Text("e-mail").Class("active deep-orange-text"),
			),

			//app.Div().Class("small-space"),

			// request button
			app.Div().Class("row center-align").Body(
				app.Button().ID("request-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickRequest).Disabled(c.buttonsDisabled).TabIndex(2).Body(
					app.Text("request"),
				),
			),

		// passphrase reset --- insert the UUID
		).Else(

			// pwd reset lightbulb
			app.Article().Class("row surface-container-highest").Body(
				app.I().Text("lightbulb").Class("amber-text"),
				app.P().Class("max").Body(
					//app.Span().Class("deep-orange-text").Text(" "),
					app.Span().Text("enter the UUID code which has been sent to your inbox; if the code is correct, your passphrase will be automatically regenerated and another confirmation mail containing the passphrase will be sent to your e-mail address"),
				),
			),
			app.Div().Class("space"),

			// pwd reset credentials fields
			app.Div().Class("field border label deep-orange-text").Body(
				app.Input().ID("uuid-input").Type("text").Required(true).TabIndex(1).Value("").OnChange(c.ValueTo(&c.uuid)).Class("active").AutoFocus(true),
				app.Label().Text("UUID").Class("active deep-orange-text"),
			),

			//app.Div().Class("small-space"),

			// pwd reset button
			app.Div().Class("row center-align").Body(
				app.Button().ID("reset-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(2).OnClick(c.onClickReset).Disabled(c.buttonsDisabled).Body(
					app.Text("reset"),
				),
			),
		),

		app.Div().Class("space"),
	)
}
