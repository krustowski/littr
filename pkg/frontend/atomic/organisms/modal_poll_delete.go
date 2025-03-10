package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

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
			return &molecules.DeleteDialog{
				ID:             "delete-modal",
				Title:          "poll deletion",
				DeleteButtonID: m.PollID,
				//
				TextBoxClass:     "row amber-border white-text border warn thicc",
				TextBoxIcon:      "warning",
				TextBoxIconClass: "amber-text",
				TextBoxText:      "Are you sure you want to delete your poll?",
				//
				ModalButtonsDisabled:     m.ModalButtonsDisabled,
				OnClickDismissActionName: m.OnClickDismissActionName,
				OnClickDeleteActionName:  m.OnClickDeleteActionName,
			}
		}),
	)
}
