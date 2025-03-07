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

	MarkupText string

	MakeSummary bool
}

func (t *TextBox) composeContentComponent() app.UI {
	if t.MakeSummary {
		return &Details{
			Limit:         40,
			Text:          t.Text,
			FormattedText: t.MarkupText,
		}
	}

	if t.MarkupText != "" {
		return &atoms.Text{
			FormattedText: t.MarkupText,
		}
	}

	return &atoms.Text{
		PlainText: t.Text,
	}
}

func (t *TextBox) Render() app.UI {
	return app.Article().Class(t.Class).Body(
		app.I().Text(t.Icon).Class(t.IconClass),

		t.composeContentComponent(),
	)
}
