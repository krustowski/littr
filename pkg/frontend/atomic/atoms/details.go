package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Details struct {
	app.Compo

	SummaryText string
	FullText    string

	SpanID                string
	OnClickSpanActionName string
}

func (d *Details) onClickText(ctx app.Context, e app.Event) {
	if d.SpanID == "" {
		return
	}

	ctx.NewActionWithValue(d.OnClickSpanActionName, d.SpanID)
}

func (d *Details) Render() app.UI {
	return app.Details().Class("max").Body(
		app.Summary().Text(d.SummaryText).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
		app.Div().Class("space"),
		app.Span().ID(d.SpanID).Class("bold").Text(d.FullText).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line").OnClick(d.onClickText),
	)
}
