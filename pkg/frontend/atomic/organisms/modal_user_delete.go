package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

type ModalUserDelete struct {
	app.Compo

	LoggedUserNickname string

	ModalShow            bool
	ModalButtonsDisabled bool

	OnClickDismissActionName       string
	OnClickDeleteAccountActionName string
}

func (m *ModalUserDelete) Render() app.UI {
	// Account deletion modal.
	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return &molecules.DeleteDialog{
				ID:             "delete-modal",
				Title:          "account deletion",
				DeleteButtonID: m.LoggedUserNickname,
				//
				TextBoxClass:     "row amber-border white-text border danger thicc",
				TextBoxIcon:      "warning",
				TextBoxIconClass: "red-text",
				TextBoxText:      "Are you sure you want to delete your account and all posted items?",
				//
				ModalButtonsDisabled:     m.ModalButtonsDisabled,
				OnClickDismissActionName: m.OnClickDismissActionName,
				OnClickDeleteActionName:  m.OnClickDeleteAccountActionName,
			}
		}),
	)
}
