package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Counter struct {
	app.Compo

	Count int64

	ID                string
	Title             string
	Icon              string
	OnClickActionName string
}

func (c *Counter) onClick(ctx app.Context, e app.Event) {
	key := e.JSValue().Get("id")

	ctx.NewActionWithValue(c.OnClickActionName, key)
}

func (c *Counter) Render() app.UI {
	return app.Div().Body(
		app.B().Title(c.Title).Text(c.Count).Class("left-padding"),
		app.Span().Title(c.Title).Class("bold").OnClick(c.onClick).ID(c.ID).Body(
			app.I().Text(c.Icon),
		),
	)
}
