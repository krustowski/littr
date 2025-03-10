package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

type ModalPostDelete struct {
	app.Compo

	PostID string

	ModalButtonsDisabled bool
	ModalShow            bool

	OnClickDismissActionName string
	OnClickDeleteActionName  string
}

func (m *ModalPostDelete) Render() app.UI {
	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return &molecules.DeleteDialog{
				ID:             "delete-modal",
				Title:          "post deletion",
				DeleteButtonID: m.PostID,
				//
				TextBoxClass:     "row amber-border white-text border warn thicc",
				TextBoxIcon:      "warning",
				TextBoxIconClass: "amber-text",
				TextBoxText:      "Are you sure you want to delete your post?",
				//
				ModalButtonsDisabled:     m.ModalButtonsDisabled,
				OnClickDismissActionName: m.OnClickDismissActionName,
				OnClickDeleteActionName:  m.OnClickDeleteActionName,
			}
		}),
	)
}
