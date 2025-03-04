package common

import (
	"strconv"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func HandleMouseEnter(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	if elem := app.Window().GetElementByID(id); !elem.IsNull() {
		elem.Get("style").Call("setProperty", "font-size", "1.2rem")
	}
}

func HandleMouseLeave(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	if elem := app.Window().GetElementByID(id); !elem.IsNull() {
		elem.Get("style").Call("setProperty", "font-size", "1rem")
	}
}

func HandleLink(ctx app.Context, a app.Action, path, pathAlt string) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host

	if _, err := strconv.ParseFloat(id, 64); err != nil {
		path = pathAlt
	}

	// Write the link to browsers's clipboard.
	navigator := app.Window().Get("navigator")
	if !navigator.IsNull() {
		clipboard := navigator.Get("clipboard")
		if !clipboard.IsNull() && !clipboard.IsUndefined() {
			clipboard.Call("writeText", scheme+"://"+host+path+id)
		}
	}

	ctx.Navigate(path + id)
}
