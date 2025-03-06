package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
)

type TextBox struct {
	app.Compo

	Class     string
	Icon      string
	IconClass string
	Text      string

	MakeSummary bool
}

func (t *TextBox) composeContent() app.UI {
	if t.MakeSummary {
		var summary string

		if len(t.Text) > 40 {
			summary = t.Text[:40] + "..."
		}

		return &atoms.Details{
			SummaryText: summary,
			FullText:    t.Text,
		}
	}

	return app.P().Class("max bold").Body(app.Span().Text((t.Text)))
}

func (t *TextBox) Render() app.UI {
	return app.Article().Class(t.Class).Body(
		app.I().Text(t.Icon).Class(t.IconClass),

		t.composeContent(),
	)
}
