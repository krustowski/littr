package atoms

import "github.com/maxence-charriere/go-app/v10/pkg/app"

type SearchBar struct {
	app.Compo

	ID string

	OnSearchActionName string
	OnSearch           app.EventHandler
}

func (s *SearchBar) onSearch(ctx app.Context, e app.Event) {
	if s.OnSearch != nil {
		s.OnSearch(ctx, e)
		return
	}

	ctx.NewActionWithValue(s.OnSearchActionName, s.ID)
}

func (s *SearchBar) Render() app.UI {
	return app.Div().Class("field prefix round fill thicc").Body(
		app.I().Class("front").Text("search"),
		//app.Input().Type("search").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),

		app.Input().ID(s.ID).Type("text").OnChange(s.onSearch).OnSearch(s.onSearch),
	)
}
