package atoms

import (
	"regexp"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type AssemblePayload struct {
	Tag     string
	Attrs   map[string]string
	Content string
}

type Text struct {
	app.Compo

	FormattedText string
	PlainText     string

	props []AssemblePayload
}

var (
	markRegex = regexp.MustCompile(`#(\w+)( [^$]*?)?#(.*?)##(\w+)#`)
	attrRegex = regexp.MustCompile(`(\w+)='(.*?)'`)
)

func (t *Text) parseMarkupAndCompose() (elems []app.UI) {
	var lastIndex int

	if t.FormattedText == "" {
		elems = append(elems, app.Span().Text(t.PlainText))
		return
	}

	for _, match := range markRegex.FindAllStringSubmatchIndex(t.FormattedText, -1) {
		start, end := match[0], match[1]

		if start > lastIndex {
			rawText := t.FormattedText[lastIndex:start]
			elems = append(elems, app.Text(rawText))
		}

		tag := t.FormattedText[match[2]:match[3]]

		var (
			attrs      = make(map[string]string)
			attrString string
			content    string
		)

		if match[4] > 0 && match[5] > 0 {
			attrString = t.FormattedText[match[4]:match[5]]
		}

		content = t.FormattedText[match[6]:match[7]]

		// Parse attributes
		attrs = make(map[string]string)
		for _, attr := range attrRegex.FindAllStringSubmatch(attrString, -1) {
			attrs[attr[1]] = attr[2]
		}

		var compo app.UI

		switch tag {
		case "bold":
			compo = app.B().Class(attrs["class"]).Text(content)

		case "break":
			compo = app.Br().Class(attrs["class"])

		case "icon":
			compo = app.I().Text(content)

		case "link":
			compo = app.A().Href(attrs["to"]).Class(attrs["class"]).Text(content)

		default:
			compo = app.Span().Text(content)
		}

		elems = append(elems, compo)
		lastIndex = end
	}

	if lastIndex < len(t.FormattedText) {
		elems = append(elems, app.Text(t.FormattedText[lastIndex:]))
	}

	return elems
}

func (t *Text) Render() app.UI {
	if t.FormattedText == "" {
		return app.P().Class("max").Text(t.PlainText)
	}

	return app.P().Class("max").Body(t.parseMarkupAndCompose()...)
}
