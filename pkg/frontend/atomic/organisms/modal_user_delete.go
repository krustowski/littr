package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

type ModalUserDelete struct {
	app.Compo

	LoggedUserNickname string

	ModalShow bool

	OnClickDismissActionName       string
	OnClickDeleteAccountActionName string
}

func (m *ModalUserDelete) Render() app.UI {
	// Account deletion modal.
	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return app.Dialog().ID("delete-modal").Class("grey10 white-text thicc active").Body(
				&atoms.PageHeading{
					Class: "center-align",
					Title: "account deletion",
					Level: 6,
				},

				&molecules.TextBox{
					Class:     "row border white-text redd-border thicc danger",
					Icon:      "warning",
					IconClass: "red-text",
					Text:      "Are you sure you want to delete your account and all posted items?",
				},
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					&atoms.Button{
						Class:             "max bold black white-text thicc",
						Icon:              "close",
						Text:              "Cancel",
						OnClickActionName: m.OnClickDismissActionName,
					},

					&atoms.Button{
						ID:                m.LoggedUserNickname,
						Class:             "max bold red10 white-text thicc",
						Icon:              "delete",
						Text:              "Delete",
						OnClickActionName: m.OnClickDismissActionName,
					},
				),
			)
		}),
	)
}
