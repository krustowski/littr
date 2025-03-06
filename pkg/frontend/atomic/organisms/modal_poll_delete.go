package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

type ModalPollDelete struct {
	app.Compo

	PollID string

	ModalButtonsDisabled bool
	ModalShow            bool

	OnClickDismissActionName string
	OnClickDeleteActionName  string
}

func (m *ModalPollDelete) Render() app.UI {
	return app.Div().Body(
		// poll deletion modal
		app.If(m.ModalShow, func() app.UI {
			return app.Div().Body(
				app.Dialog().ID("delete-modal").Class("grey10 white-text active thicc").Body(
					app.Nav().Class("center-align").Body(
						app.H5().Text("poll deletion"),
					),

					app.Div().Class("space"),

					&molecules.TextBox{
						Class:     "row border amber-border white-text warn thicc",
						IconClass: "amber-text",
						Icon:      "warning",
						Text:      "Are you sure you want to delete your poll?",
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
							ID:                m.PollID,
							Class:             "max bold red10 white-text thicc",
							Icon:              "delete",
							Text:              "Delete",
							OnClickActionName: m.OnClickDeleteActionName,
							Disabled:          m.ModalButtonsDisabled,
						},
					),
				),
			)
		}),
	)
}
