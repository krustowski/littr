package login

import (
	"go.vxn.dev/littr/pkg/config"

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

		// Login credentials fields
		app.Div().Class("field border label deep-orange-text").Style("border-radius", "8px").Body(
			app.Input().ID("login-input").Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(config.NicknameLengthMax).Class("active").Attr("autocomplete", "username"),
			app.Label().Text("Nickname").Class("active deep-orange-text"),
		),

		app.Div().Class("field border label deep-orange-text").Style("border-radius", "8px").Body(
			app.Input().ID("passphrase-input").Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Class("active").Attr("autocomplete", "current-password"),
			app.Label().Text("Passphrase").Class("active deep-orange-text"),
		),

		// Session duration infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Text("The login session lasts 30 days."),
			),
		),
		app.Div().Class("space"),

		// login button
		app.Div().Class("row center-align").Body(
			app.Button().ID("login-button").Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClick).Disabled(c.loginButtonDisabled).TabIndex(3).Body(
				app.If(c.loginButtonDisabled,
					app.Progress().Class("circle white-border small"),
				),
				app.Span().Body(
					app.I().Style("padding-right", "5px").Text("login"),
					app.Text("Login"),
				),
			),
		),
		app.Div().Class("space"),

		// reset button
		app.Div().Class("row center-align").Body(
			app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(4).OnClick(c.onClickReset).Disabled(c.loginButtonDisabled).Body(
				app.Span().Body(
					app.I().Style("padding-right", "5px").Text("password"),
					app.Text("Recover passphrase"),
				),
			),
		),
		app.Div().Class("space"),

		// register button
		app.Div().Class("row center-align").Body(
			// register button
			app.If(config.IsRegistrationEnabled,
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(5).OnClick(c.onClickRegister).Disabled(c.loginButtonDisabled).Body(
					app.Span().Body(
						app.I().Style("padding-right", "5px").Text("app_registration"),
						app.Text("Register"),
					),
				),
			).Else(
				app.Button().Class("max shrink deep-orange7 white-text bold").Style("border-radius", "8px").TabIndex(5).OnClick(nil).Disabled(true).Body(
					app.Span().Body(
						app.I().Style("padding-right", "5px").Text("app_registration"),
						app.Text("Registration disabled"),
					),
				),
			),
		),
		app.Div().Class("medium-space"),
	)
}
