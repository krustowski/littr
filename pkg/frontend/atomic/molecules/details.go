package molecules

import (
	"regexp"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
)

type Details struct {
	app.Compo

	Limit int

	Text          string
	FormattedText string

	SpanID                string
	OnClickSpanActionName string
}

func (d *Details) onClickText(ctx app.Context, e app.Event) {
	if d.SpanID == "" {
		return
	}

	ctx.NewActionWithValue(d.OnClickSpanActionName, d.SpanID)
}

func (d *Details) stripMarkup() string {
	// Match tags and extract content only
	tagRegex := regexp.MustCompile(`#(\w+)( [^$]*?)?#(.*?)##(\w+)#`)
	plainText := tagRegex.ReplaceAllString(d.FormattedText, "$3") // Keep only inner content

	// Trim excessive spaces
	return strings.TrimSpace(plainText)
}

func (d *Details) Render() app.UI {
	// Limited text summary.
	summaryBody := func() app.UI {
		if d.FormattedText != "" {
			return app.Text(d.stripMarkup()[:d.Limit] + "...")
		}

		if len(d.Text) < d.Limit {
			return app.Text(d.Text)
		}

		return app.Text(d.Text[:d.Limit] + "...")
	}

	// Full text span body.
	spanBody := func() app.UI {
		if d.FormattedText != "" {
			return &atoms.Text{
				FormattedText: d.FormattedText,
			}
		}

		return app.Text(d.Text)
	}

	return app.Details().Class("max").Body(
		app.Summary().Style("word-break", "break-word").Style("hyphens", "auto").Class("italic").Body(summaryBody()),
		app.Div().Class("space"),
		app.Span().ID(d.SpanID).Class("").Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line").OnClick(d.onClickText).Body(spanBody()),
	)
}
