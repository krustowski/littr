package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
)

type Switch struct {
	app.Compo

	Icon string
	ID   string
	Text string

	Checked  bool
	Disabled bool

	OnChangeActionName string
}

func (s *Switch) Render() app.UI {
	return app.Div().Class("field middle-align").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max").Body(
				app.Span().Text(s.Text),
			),
			app.Label().Class("switch icon").Body(
				&atoms.Input{
					ID:                 s.ID,
					Type:               "checkbox",
					Checked:            s.Checked,
					Disabled:           s.Disabled,
					OnChangeType:       atoms.InputOnChangeEventHandler,
					OnChangeActionName: s.OnChangeActionName,
				},
				app.Span().Body(
					app.I().Text(s.Icon),
				),
			),
		),
	)
}
