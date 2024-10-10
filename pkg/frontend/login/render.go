package login

import (
	"go.vxn.dev/littr/configs"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("littr login"),
			),
		),
		app.Div().Class("space"),

		// snackbar
		app.A().Href(c.toast.TLink).OnClick(c.onDismissToast).Body(
			app.If(c.toast.TText != "",
				app.Div().ID("snackbar").Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				),
			),
		),

		// login credentials fields
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("login-input").Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(configs.NicknameLengthMax).Class("active").Attr("autocomplete", "username"),
			app.Label().Text("nickname").Class("active deep-orange-text"),
		),

		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("passphrase-input").Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Class("active").Attr("autocomplete", "current-password"),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),
		app.Article().Class("row surface-container-highest").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text(" "),
				app.Span().Text("log-in for 30 days"),
			),
		),
		app.Div().Class("space"),

		// login button
		app.Div().Class("row center-align").Body(
			app.Button().ID("login-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(3).OnClick(c.onClick).Disabled(c.loginButtonDisabled).Body(
				app.Text("login"),
			),
		),
		app.Div().Class("space"),

		// reset button
		app.Div().Class("row center-align").Body(
			app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(4).OnClick(c.onClickReset).Disabled(c.loginButtonDisabled).Body(
				app.Text("recover forgotten passphrase"),
			),
		),
		app.Div().Class("space"),

		// register button
		app.Div().Class("row center-align").Body(
			// register button
			app.If(app.Getenv("REGISTRATION_ENABLED") == "true",
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(5).OnClick(c.onClickRegister).Disabled(c.loginButtonDisabled).Body(
					app.Text("register"),
				),
			).Else(
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(5).OnClick(nil).Disabled(true).Body(
					app.Text("register"),
				),
			),
		),
		app.Div().Class("medium-space"),
	)
}
