package molecules

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type LittrHeader struct {
	app.Compo

	HeaderString string

	OnClickHeadlineActionName string
}

func (l *LittrHeader) onClick(ctx app.Context, e app.Event) {
	id := e.JSValue().Get("id").String()

	ctx.NewActionWithValue(l.OnClickHeadlineActionName, id)
}

func (l *LittrHeader) Render() app.UI {
	// littr header
	return app.H4().Title("system info (click to open)").Class("center-align blue-text").OnClick(l.onClick).ID("top-header").Body(
		app.Span().Body(
			app.Text(l.HeaderString),

			app.If(app.Getenv("APP_ENVIRONMENT") != "prod", func() app.UI {
				return app.Span().Class("col").Body(
					app.Sup().Body(
						app.If(app.Getenv("APP_ENVIRONMENT") == "stage", func() app.UI {
							return app.Text(" (stage) ")
						}).Else(func() app.UI {
							return app.Text(" (dev) ")
						}),
					),
				)
			}),
		),
	)
}
