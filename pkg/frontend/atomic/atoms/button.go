package atoms

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Button struct {
	app.Compo

	Class             string
	Color             string
	ColorText         string
	Icon              string
	ID                string
	Name              string
	Text              string
	Title             string
	OnClickActionName string

	DataSet map[string]string

	Attr map[string]string

	Disabled     bool
	ShowProgress bool

	OnClick app.EventHandler
}

func (b *Button) onClick(ctx app.Context, e app.Event) {
	if b.OnClick != nil {
		b.OnClick(ctx, e)
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		b.Disabled = true
		b.ShowProgress = true
	})

	ctx.Defer(func(ctx app.Context) {
		ctx.NewActionWithValue(b.OnClickActionName, b.ID)
	})
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

	for key, val := range b.DataSet {
		bt.DataSet(key, val)
	}

	return bt.ID(b.ID).Name(b.Name).Title(b.Title).Class(b.composeClass()).OnClick(b.onClick).Disabled(b.Disabled).Body(
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
