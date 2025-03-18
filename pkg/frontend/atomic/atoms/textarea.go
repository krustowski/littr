package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type Textarea struct {
	app.Compo

	ID        string
	Class     string
	Content   string
	Name      string
	LabelText string

	ContentPointer *string

	OnBlurActionName string
}

func (t *Textarea) onBlur(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(t.OnBlurActionName, e.Get("id").String())
}

func (t *Textarea) OnMount(ctx app.Context) {
	if t.ContentPointer == nil {
		t.ContentPointer = new(string)
	}

	if t.Content == "" {
		t.Content = *t.ContentPointer
	}
}

func (t *Textarea) Render() app.UI {
	return app.Div().Class(t.Class).Style("border-radius", "8px").Body(
		app.Textarea().Class("primary-text active").Name(t.Name).Text(t.Content).OnChange(t.ValueTo(t.ContentPointer)).AutoFocus(true).ID(t.ID).OnBlur(t.onBlur),
		app.Label().Text(t.LabelText).Class("active primary-text"),
	)
}
