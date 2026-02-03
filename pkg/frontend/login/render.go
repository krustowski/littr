package login

import (
	"go.vxn.dev/littr/pkg/config"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive shrink-20").Body(
		app.Div().Class("row left-align").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("login"),
			),
		),
		app.Div().Class("space"),

		// Login credentials fields
		app.Div().Class("field border label primary-text thicc center-align").Body(
			app.Input().ID("login-input").Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(config.MaxNicknameLength).Class("active").Attr("autocomplete", "username"),
			app.Label().Text("Nickname").Class("active primary-text"),
		),

		app.Div().Class("field border label primary-text thicc center-align").Body(
			app.Input().ID("passphrase-input").Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Class("active").Attr("autocomplete", "current-password"),
			app.Label().Text("Passphrase").Class("active primary-text"),
		),

		// Session duration infobox.
		app.Article().Class("row border blue-border info thicc").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				app.Span().Text("The login session lasts 30 days."),
			),
		),
		app.Div().Class("space"),

		// login button
		app.Div().Class("row center-align").Body(
			app.Button().ID("login-button").Class("max primary-container white-text bold thicc shrink").OnClick(c.onClick).Disabled(c.loginButtonDisabled).TabIndex(3).Body(
				app.If(c.loginButtonDisabled, func() app.UI {
					return app.Progress().Class("circle white-border small")
				}),
				app.Span().Body(
					app.I().Style("padding-right", "5px").Text("login"),
					app.Text("Login"),
				),
			),
		),
		app.Div().Class("little-space"),

		// reset button
		app.Div().Class("row center-align").Body(
			app.Button().Class("max primary-container white-text bold thicc shrink").TabIndex(4).OnClick(c.onClickReset).Disabled(c.loginButtonDisabled).Body(
				app.Span().Body(
					app.I().Style("padding-right", "5px").Text("password"),
					app.Text("Recover"),
				),
			),
		),
		app.Div().Class("little-space"),

		// register button
		app.Div().Class("row center-align").Body(
			// register button
			app.If(config.IsRegistrationEnabled, func() app.UI {
				return app.Button().Class("max primary-container white-text bold thicc shrink").TabIndex(5).OnClick(c.onClickRegister).Disabled(c.loginButtonDisabled).Body(
					app.Span().Body(
						app.I().Style("padding-right", "5px").Text("app_registration"),
						app.Text("Register"),
					),
				)
			}).Else(func() app.UI {
				return app.Button().Class("max primary-container white-text bold thicc shrink").TabIndex(5).OnClick(nil).Disabled(true).Body(
					app.Span().Body(
						app.I().Style("padding-right", "5px").Text("app_registration"),
						app.Text("Registration disabled"),
					),
				)
			}),
		),
		app.Div().Class("medium-space"),
	)
}
