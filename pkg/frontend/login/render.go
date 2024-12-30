package login

import (
	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/common"

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
				app.Div().ID("snackbar").Class("snackbar white-text top active "+common.ToastColor(c.toast.TType)).Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				),
			),
		),

		// login credentials fields
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("login-input").Type("text").Required(true).TabIndex(1).OnChange(c.ValueTo(&c.nickname)).MaxLength(config.NicknameLengthMax).Class("active").Attr("autocomplete", "username"),
			app.Label().Text("nickname").Class("active deep-orange-text"),
		),

		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("passphrase-input").Type("password").Required(true).TabIndex(2).OnChange(c.ValueTo(&c.passphrase)).MaxLength(50).Class("active").Attr("autocomplete", "current-password"),
			app.Label().Text("passphrase").Class("active deep-orange-text"),
		),

		// Session duration infobox.
		app.Article().Class("row surface-container-highest").Style("border-radius", "8px").Body(
			app.I().Text("info").Class("blue-text"),
			app.P().Class("max").Body(
				//app.Span().Class("deep-orange-text").Text(" "),
				app.Span().Text("The login session lasts 30 days."),
			),
		),
		app.Div().Class("space"),

		// login button
		app.Div().Class("row center-align").Body(
			app.Button().ID("login-button").Class("max shrink center deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClick).Disabled(c.loginButtonDisabled).TabIndex(3).Body(
				app.If(c.loginButtonDisabled,
					app.Progress().Class("circle white-border small"),
				),
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
			app.If(config.IsRegistrationEnabled,
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
