package register

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("littr registration"),
			),
		),
		app.Div().Class("space"),

		// snackbar
		app.A().Href(t.toast.TLink).OnClick(c.onDismissToast).Body(
			app.If(c.toast.TText != "",
				app.Div().ID("snackbar").Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				),
			),
		),

		// nickname field
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("nickname "),
				app.Span().Text("is your unique identifier for the system operations; "),
				app.Span().Text("please double-check your nickname before registering (nickname is case-sensitive)"),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("nickname-input").Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true).Class("active").AutoFocus(true).MaxLength(50).Attr("autocomplete", "username").TabIndex(1).Name("login"),
			app.Label().Text("nickname").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// password fields
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("passphrase "),
				app.Span().Text("is your secret key to the littr account"),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-input").Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(2),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),
		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("passphrase-again-input").Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(3),
			app.Label().Text("passphrase again").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// e-mail field
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("e-mail "),
				app.Span().Text("address is used for user's avatar fetching from Gravatar.com, and (not yet implemented) for the account verification, please enter a valid e-mail address"),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Body(
			app.Input().ID("email-input").Type("email").OnChange(c.ValueTo(&c.email)).Required(true).Class("active").MaxLength(60).Attr("autocomplete", "email").TabIndex(4),
			app.Label().Text("e-mail").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// GDPR warning
		app.Article().Class("row surface-container-highest").Style("word-break", "break-word").Body(
			app.I().Text("warning").Class("red-text"),
			app.Div().Class("max").Style("word-break", "break-word").Style("hyphens", "auto").Body(
				app.P().Style("word-break", "break-word").Style("hyphens", "auto").Body(
					app.Span().Text("by clicking on the register button you are giving us a GDPR consent (a permission to store your account information in the database)"),
				),
				app.P().Text("you can flush your account data and published posts simply on the settings page after a log-in"),
			),
		),
		app.Div().Class("space"),

		// register button
		app.Div().Class("row center-align").Body(
			app.If(app.Getenv("REGISTRATION_ENABLED") == "true",
				app.Button().ID("register-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickRegister).Disabled(c.registerButtonDisabled).TabIndex(5).Body(
					app.Text("register"),
				),
			).Else(
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(nil).Disabled(true).Body(
					app.Text("registration off"),
				),
			),
		),

		app.Div().Class("medium-space"),
	)
}
