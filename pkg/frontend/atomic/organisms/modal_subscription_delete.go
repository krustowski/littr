package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

type ModalSubscriptionDelete struct {
	app.Compo

	ModalShow bool

	OnClickDismissActionName string
	OnClickDeleteActionName  string
}

func (m *ModalSubscriptionDelete) Render() app.UI {
	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return app.Dialog().ID("delete-modal").Class("grey10 white-text active thicc").Body(
				&atoms.PageHeading{
					Class: "center-align",
					Title: "subscription deletion",
					Level: 6,
				},

				&molecules.TextBox{
					Class:     "row border white-text amber-border thicc danger",
					Icon:      "warning",
					IconClass: "amber-text",
					Text:      "Are you sure you want to delete this subscription?",
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
						Class:             "max bold red10 white-text thicc",
						Icon:              "delete",
						Text:              "Delete",
						OnClickActionName: m.OnClickDeleteActionName,
					},
				),
			)
		}),
	)
}
