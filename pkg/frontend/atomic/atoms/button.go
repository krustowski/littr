package atoms

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Button struct {
	app.Compo

	Class     string
	Color     string
	ColorText string
	Icon      string
	ID        string
	Text      string
	Title     string

	Attr map[string]string

	Disabled     bool
	ShowProgress bool

	OnClick app.EventHandler
}

func (b *Button) composeClass() string {
	if b.Class != "" {
		return b.Class
	}

	return fmt.Sprintf("max shrink %s %s bold thicc", b.Color, b.ColorText)
}

func (b *Button) Render() app.UI {
	bt := app.Button()

	for key, val := range b.Attr {
		bt.Attr(key, val)
	}

	return bt.ID(b.ID).Title(b.Title).Class(b.composeClass()).OnClick(b.OnClick).Disabled(b.Disabled).Body(
		app.If(b.Disabled && b.ShowProgress, func() app.UI {
			return app.Progress().Class("circle white-border small")
		}),
		app.If(b.Text != "", func() app.UI {
			return app.Span().Body(
				app.I().Style("padding-right", "5px").Text(b.Icon),
				app.Text(b.Text),
			)
		}).Else(func() app.UI {
			return app.I().Text(b.Icon)
		}),
	)
}
