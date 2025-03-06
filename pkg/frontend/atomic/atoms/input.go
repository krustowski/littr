package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type onChangeType byte

const (
	InputOnChangeEventHandler onChangeType = iota
	InputOnChangeValueTo
)

type Input struct {
	app.Compo

	ID    string
	Type  string
	Class string
	Value string

	MaxLength int

	AutoComplete bool
	Checked      bool
	Disabled     bool

	OnChangeType       onChangeType
	OnChangeActionName string
}

func (i *Input) onChange(ctx app.Context, e app.Event) {
	if i.ID == "" || i.OnChangeActionName == "" {
		return
	}

	ctx.NewActionWithValue(i.OnChangeActionName, i.ID)
}

func (i *Input) Render() app.UI {
	ipt := app.Input()

	// OnChange(<eventHandler>)
	// OnChange(compo.ValueTo(compo.Content))
	switch i.OnChangeType {
	case InputOnChangeEventHandler:
		ipt.OnChange(i.onChange)
	case InputOnChangeValueTo:
		ipt.OnChange(i.ValueTo(&i.Value))
	}

	return ipt.Class(i.Class).Type(i.Type).ID(i.ID).Checked(i.Checked).Disabled(i.Disabled).AutoComplete(i.AutoComplete).MaxLength(i.MaxLength).Value(i.Value)
}
