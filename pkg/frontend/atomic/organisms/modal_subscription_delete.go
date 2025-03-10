package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

type ModalSubscriptionDelete struct {
	app.Compo

	ModalShow            bool
	ModalButtonsDisabled bool

	OnClickDismissActionName string
	OnClickDeleteActionName  string
}

func (m *ModalSubscriptionDelete) Render() app.UI {
	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return &molecules.DeleteDialog{
				ID:    "delete-modal",
				Title: "subscription deletion",
				//
				TextBoxClass:     "row amber-border white-text border warn thicc",
				TextBoxIcon:      "warning",
				TextBoxIconClass: "amber-text",
				TextBoxText:      "Are you sure you want to delete this subscription?",
				//
				ModalButtonsDisabled:     m.ModalButtonsDisabled,
				OnClickDismissActionName: m.OnClickDismissActionName,
				OnClickDeleteActionName:  m.OnClickDeleteActionName,
			}
		}),
	)
}
