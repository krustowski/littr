package common

import "github.com/maxence-charriere/go-app/v10/pkg/app"

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
