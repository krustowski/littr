package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type PollHeader struct {
	app.Compo

	Poll models.Poll

	ButtonsDisabled bool

	OnClickLinkActionName string
}

func (p *PollHeader) Render() app.UI {
	return app.Div().Class("row top-padding bottom-padding").Body(

		app.P().Class("max").Body(
			app.Span().Title("question").Text("Q: "),
			app.Span().Text(p.Poll.Question).Class("deep-orange-text space bold"),
		),

		&atoms.Button{
			ID:                p.Poll.ID,
			Title:             "link to this poll",
			Class:             "transparent circle",
			Icon:              "link",
			OnClickActionName: p.OnClickLinkActionName,
			Disabled:          p.ButtonsDisabled,
		},
	)
}
