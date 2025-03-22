package register

import (
	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		&atoms.PageHeading{
			Title: "registration",
			Level: 5,
		},
		app.Div().Class("space"),

		// Nickname field.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  "#bold class='blue-text'#Nickname##bold# is your unique identifier on the site. Please double-check the input before submitting (nickname is case-sensitive).",
			MakeSummary: false,
		},
		app.Div().Class("space"),

		app.Div().Class("field label border primary-text thicc").Body(
			&atoms.Input{
				ID:           "nickname-input",
				Type:         "text",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.nickname,
				AutoComplete: true,
				MaxLength:    config.MaxNicknameLength,
				Attr: map[string]string{
					"autocomplete": "username",
				},
				//Required: true,
				//TabIndex: 1,
			},
			app.Label().Text("Nickname").Class("active primary-text"),
		),
		app.Div().Class("space"),

		// Passphrase fields.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  "#bold class='blue-text'#Passphrase##bold# is a secret key to your account. Keep it strong and private.",
			MakeSummary: false,
		},
		app.Div().Class("space"),

		app.Div().Class("field label border primary-text thicc").Body(
			&atoms.Input{
				ID:           "passphrase-input",
				Type:         "password",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.passphrase,
				AutoComplete: true,
				Attr: map[string]string{
					"autocomplete": "new-password",
				},
				//MaxLength: 50,
				//Required: true,
				//TabIndex: 2,
			},
			app.Label().Text("Passphrase").Class("active primary-text"),
		),
		app.Div().Class("field label border primary-text thicc").Body(
			&atoms.Input{
				ID:           "passphrase-again-input",
				Type:         "password",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.passphraseAgain,
				AutoComplete: true,
				Attr: map[string]string{
					"autocomplete": "new-password",
				},
				//MaxLength: 50,
				//Required: true,
				//TabIndex: 3,
			},
			app.Label().Text("Passphrase again").Class("active primary-text"),
		),
		app.Div().Class("space"),

		// E-mail field.
		&molecules.TextBox{
			Class:       "row border blue-border thicc info",
			Icon:        "info",
			IconClass:   "blue-text",
			MarkupText:  "#bold class='blue-text'#E-mail##bold# address is used for the account verification. You will receive a link to activate your account easily. Also, if the address is registered with a Gravatar account, such avatar will be loaded from there. Please enter a valid e-mail address.",
			MakeSummary: false,
		},
		app.Div().Class("space"),

		app.Div().Class("field label border primary-text thicc").Body(
			&atoms.Input{
				ID:           "email-input",
				Type:         "email",
				Class:        "active",
				OnChangeType: atoms.InputOnChangeValueTo,
				Content:      c.email,
				AutoComplete: true,
				MaxLength:    config.MaxNicknameLength,
				Attr: map[string]string{
					"autocomplete": "email",
				},
				//Required: true,
				//TabIndex: 4,
			},
			app.Label().Text("E-mail").Class("active primary-text"),
		),
		app.Div().Class("space"),

		// GDPR notice.
		&molecules.TextBox{
			Class:       "row border amber-border thicc warn",
			Icon:        "warning",
			IconClass:   "amber-text",
			MarkupText:  "By clicking on the register button you are giving us a GDPR consent (a permission to store your account information in the database).#break###break #break###break# You can flush your account's data and published posts/polls simply on the settings page after a log-in. Any time.",
			MakeSummary: false,
		},
		app.Div().Class("space"),

		// Register button.
		app.Div().Class("row center-align").Body(
			app.If(config.IsRegistrationEnabled, func() app.UI {
				return app.Div().Class("row max").Body(
					&atoms.Button{
						ID:       "register-button",
						Class:    "max shrink center primary-container bold thicc",
						OnClick:  c.onClickRegister,
						Disabled: c.registerButtonDisabled,
						Icon:     "app_registration",
						Text:     "Register",
						//TabIndex: 5
					},
				)
			}).Else(func() app.UI {
				return app.Div().Class("row max").Body(
					&atoms.Button{
						Class:    "max shrink center primary-container bold thicc",
						Disabled: true,
						Icon:     "close",
						Text:     "Registration disabled",
						//TabIndex: 5
					},
				)
			}),
		),

		app.Div().Class("medium-space"),
	)
}
