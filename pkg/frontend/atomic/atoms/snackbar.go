package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Snackbar struct {
	app.Compo

	Class    string
	ID       string
	IDLink   string
	Position string
	Text     string

	Styles map[string]string
}

func (s *Snackbar) Render() app.UI {
	return app.A().ID(s.IDLink).Href("").Body(
		app.If(s.Text != "", func() app.UI {
			sb := app.Div()

			sb.Class(s.Position)

			for key, val := range s.Styles {
				sb.Style(key, val)
			}

			return sb.ID(s.ID).Class(s.Class).Body()
		}),
	)
}
