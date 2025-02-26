package atoms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type UserNickname struct {
	app.Compo

	Class                  string
	Nickname               string
	SpanID                 string
	Title                  string
	Text                   string
	OnClickActionName      string
	OnMouseEnterActionName string
	OnMouseLeaveActionName string
}

func (u *UserNickname) onClick(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(u.OnClickActionName, u.Nickname)
}

func (u *UserNickname) onMouseEnter(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(u.OnMouseEnterActionName, u.SpanID)
}

func (u *UserNickname) onMouseLeave(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(u.OnMouseLeaveActionName, u.SpanID)
}

func (u *UserNickname) Render() app.UI {
	return app.P().Class("max").Body(
		app.A().ID(u.Nickname).Title(u.Title).Class(u.Class).OnClick(u.onClick).Body(
			app.Span().ID(u.SpanID).Class(u.Class).Text(u.Nickname).OnMouseEnter(u.onMouseEnter).OnMouseLeave(u.onMouseLeave),
		),
	)
}
