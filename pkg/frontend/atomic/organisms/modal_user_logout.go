package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type ModalUserLogout struct {
	app.Compo

	User models.User

	ShowModal bool

	OnClickDismissActionName string
	OnClickLogoutActionName  string
	OnClickFlowActionName    string
}

func (m *ModalUserLogout) Render() app.UI {
	return app.Div().Body(
		app.If(m.ShowModal, func() app.UI {
			return app.Dialog().ID("logout-modal").Class("grey10 white-text active thicc").Body(
				&atoms.PageHeading{
					Title: "user",
					Class: "max center-align",
				},

				// User's avatar and nickname.
				app.Div().Class("row border thicc").Body(
					&atoms.Image{
						ID:                m.User.Nickname,
						Title:             "user's flow link",
						Src:               m.User.AvatarURL,
						Class:             "responsive padding max",
						OnClickActionName: m.OnClickFlowActionName,
						Styles:            map[string]string{"max-height": "100%", "max-width": "10rem", "border-radius": "50%"},
					},

					&atoms.Button{
						ID:                m.User.Nickname,
						Class:             "max bold primary-container white-text right-margin thicc",
						Icon:              "tsunami",
						Text:              "Flow",
						OnClickActionName: m.OnClickFlowActionName,
					},
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					&atoms.Button{
						Class:             "max bold black white-text thicc",
						Icon:              "close",
						Text:              "Close",
						OnClickActionName: m.OnClickDismissActionName,
					},

					&atoms.Button{
						Class:             "max primary-container white-text thicc",
						Icon:              "logout",
						Text:              "Log out",
						OnClickActionName: m.OnClickLogoutActionName,
					},
				),
			)
		}),
	)
}
