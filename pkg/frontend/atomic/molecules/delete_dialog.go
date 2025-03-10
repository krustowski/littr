package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
)

type DeleteDialog struct {
	app.Compo

	ID    string
	Title string

	TextBoxClass     string
	TextBoxIcon      string
	TextBoxIconClass string
	TextBoxText      string

	DeleteButtonID string

	ModalButtonsDisabled bool

	OnClickDismissActionName string
	OnClickDeleteActionName  string
}

func (d *DeleteDialog) Render() app.UI {
	return app.Dialog().ID(d.ID).Class("grey10 white-text active thicc").Body(
		&atoms.PageHeading{
			Class: "center",
			Title: d.Title,
		},
		app.Div().Class("space"),

		&TextBox{
			Class:     d.TextBoxClass,
			IconClass: d.TextBoxIconClass,
			Icon:      d.TextBoxIcon,
			Text:      d.TextBoxText,
		},
		app.Div().Class("space"),

		app.Div().Class("row").Body(
			&atoms.Button{
				Class:             "max bold black white-text thicc",
				Icon:              "close",
				Text:              "Cancel",
				OnClickActionName: d.OnClickDismissActionName,
				Disabled:          d.ModalButtonsDisabled,
			},
			&atoms.Button{
				ID:                d.DeleteButtonID,
				Class:             "max bold red10 white-text thicc",
				Icon:              "delete",
				Text:              "Delete",
				OnClickActionName: d.OnClickDeleteActionName,
				Disabled:          d.ModalButtonsDisabled,
			},
		),
	)
}
