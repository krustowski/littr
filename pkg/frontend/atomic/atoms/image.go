package atoms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Image struct {
	app.Compo

	ID    string
	Title string
	Class string
	Src   string

	Attr   map[string]string
	Styles map[string]string

	OnClick           app.EventHandler
	OnClickActionName string
}

func (i *Image) onClick(ctx app.Context, e app.Event) {
	if i.OnClick != nil {
		i.OnClick(ctx, e)
		return
	}

	ctx.NewActionWithValue(i.OnClickActionName, e.Get("id").String())
}

func (i *Image) Render() app.UI {
	img := app.Img()

	for key, val := range i.Attr {
		img.Attr(key, val)
	}

	for key, val := range i.Styles {
		img.Style(key, val)
	}

	return img.ID(i.ID).Title(i.Title).Class(i.Class).Src(i.Src).OnClick(i.onClick)
}
