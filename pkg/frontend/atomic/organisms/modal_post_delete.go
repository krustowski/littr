package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
)

type ModalPostDelete struct {
	app.Compo

	ModalButtonsDisabled bool
	ModalShow            bool

	OnClickDismissActionName string
	OnClickDeleteActionName  string
}

func (m *ModalPostDelete) Render() app.UI {
	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return app.Dialog().ID("delete-modal").Class("grey10 white-text active thicc").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("post deletion"),
				),
				app.Div().Class("space"),

				&atoms.TextBox{
					Class:     "row amber-border white-text border warn thicc",
					IconClass: "amber-text",
					Icon:      "warning",
					Text:      "Are you sure you want to delete your post?",
				},
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					&atoms.Button{
						Class:             "max bold black white-text thicc",
						Icon:              "close",
						Text:              "Cancel",
						OnClickActionName: m.OnClickDismissActionName,
						Disabled:          m.ModalButtonsDisabled,
					},
					&atoms.Button{
						Class:             "max bold red10 white-text thicc",
						Icon:              "delete",
						Text:              "Delete",
						OnClickActionName: m.OnClickDeleteActionName,
						Disabled:          m.ModalButtonsDisabled,
					},
				),
			)
		}),
	)
}
