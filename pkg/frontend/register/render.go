package register

import (
	"go.vxn.dev/littr/pkg/config"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("registration"),
			),
		),
		app.Div().Class("space"),

		// Nickname field.
		app.Article().Class("row border blue-border surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("Nickname "),
				app.Span().Text("is your unique identifier on the site. Please double-check the input before registering (nickname is case-sensitive)."),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Style("border-radius", "8px").Body(
			app.Input().ID("nickname-input").Type("text").OnChange(c.ValueTo(&c.nickname)).Required(true).Class("active").AutoFocus(true).MaxLength(50).Attr("autocomplete", "username").TabIndex(1).Name("login"),
			app.Label().Text("Nickname").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// Passphrase fields.
		app.Article().Class("row border blue-border surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("Passphrase "),
				app.Span().Text("is a secret key to your account. Keep it strong and private."),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Style("border-radius", "8px").Body(
			app.Input().ID("passphrase-input").Type("password").OnChange(c.ValueTo(&c.passphrase)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(2),
			app.Label().Text("Passphrase").Class("active deep-orange-text"),
		),
		app.Div().Class("field label border deep-orange-text").Style("border-radius", "8px").Body(
			app.Input().ID("passphrase-again-input").Type("password").OnChange(c.ValueTo(&c.passphraseAgain)).Required(true).Class("active").MaxLength(50).Attr("autocomplete", "new-password").TabIndex(3),
			app.Label().Text("Passphrase again").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// E-mail field.
		app.Article().Class("row border blue-border surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("E-mail "),
				app.Span().Text("address is used for the account verification. You will receive a link to activate your account easily. Also, if the address is registered with a Gravatar account, such avatar will be loaded from there. Please enter a valid e-mail address."),
			),
		),
		app.Div().Class("space"),

		app.Div().Class("field label border deep-orange-text").Style("border-radius", "8px").Body(
			app.Input().ID("email-input").Type("email").OnChange(c.ValueTo(&c.email)).Required(true).Class("active").MaxLength(60).Attr("autocomplete", "email").TabIndex(4),
			app.Label().Text("E-mail").Class("active deep-orange-text"),
		),
		app.Div().Class("space"),

		// GDPR notice.
		app.Article().Class("row border amber-border surface-container-highest").Style("border-radius", "8px").Style("word-break", "break-word").Body(
			app.I().Text("warning").Class("amber-text"),
			app.Div().Class("max").Style("word-break", "break-word").Style("hyphens", "auto").Body(
				app.P().Style("word-break", "break-word").Style("hyphens", "auto").Body(
					app.Span().Text("By clicking on the register button you are giving us a GDPR consent (a permission to store your account information in the database)."),
				),
				app.P().Text("You can flush your account's data and published posts/polls simply on the settings page after a log-in. Any time."),
			),
		),
		app.Div().Class("space"),

		// Register button.
		app.Div().Class("row center-align").Body(
			app.If(config.IsRegistrationEnabled,
				app.Div().Class("row max").Body(
					app.Button().ID("register-button").Class("max shrink center deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickRegister).Disabled(c.registerButtonDisabled).TabIndex(5).Body(
						app.If(c.registerButtonDisabled,
							app.Progress().Class("circle white-border small"),
						),
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("app_registration"),
							app.Text("Register"),
						),
					),
				),
			).Else(
				app.Div().Class("row max").Body(
					app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(nil).Disabled(true).Body(
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("app_registration"),
							app.Text("Registration disabled"),
						),
					),
				),
			),
		),

		app.Div().Class("medium-space"),
	)
}
