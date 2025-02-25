package atoms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Image struct {
	app.Compo

	ID     string
	Title  string
	Class  string
	Src    string
	Width  string
	Height string
	Radius string

	Attr map[string]string

	OnClick app.EventHandler
}

func (i *Image) composeHeight() string {
	if i.Height != "" {
		return i.Height
	}

	return "100%"
}

func (i *Image) composeRadius() string {
	if i.Radius != "" {
		return i.Radius
	}

	return "100%"
}

func (i *Image) composeWidth() string {
	if i.Width != "" {
		return i.Width
	}

	return "100%"
}

func (i *Image) Render() app.UI {
	img := app.Img()

	for key, val := range i.Attr {
		img.Attr(key, val)
	}

	return img.ID(i.ID).Title(i.Title).Class(i.Class).Src(i.Src).Style("max-height", i.composeHeight()).Style("max-width", i.composeWidth()).Style("border-radius", i.composeRadius()).OnClick(i.OnClick)
}
