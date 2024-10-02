package frontend

import (
	"go.vxn.dev/littr/pkg/frontend/polls"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

//
//  polls view
//

type PollsView struct {
	app.Compo
}

func (p *PollsView) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("polls / littr")
}

func (p *PollsView) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&polls.Content{},
	)
}
