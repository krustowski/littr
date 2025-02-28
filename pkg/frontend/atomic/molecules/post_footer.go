package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type PostFooter struct {
	app.Compo

	LoggedUserNickname string
	PostTimestamp      string

	Post models.Post

	ButtonsDisabled bool

	OnClickDeleteActionName string
	OnClickStarActionName   string
	OnClickReplyActionName  string
}

func (p *PostFooter) Render() app.UI {
	// post footer (timestamp + reply buttom + star/delete button)
	return app.Div().Class("row").Body(
		app.Div().Class("max").Body(
			//app.Text(post.Timestamp.Format("Jan 02, 2006 / 15:04:05")),
			app.Text(p.PostTimestamp),
		),

		app.If(p.Post.Nickname != "system", func() app.UI {
			return app.Div().Body(
				app.If(p.Post.ReplyCount > 0, func() app.UI {
					return app.B().Title("reply count").Text(p.Post.ReplyCount).Class("left-padding")
				}),

				&atoms.Button{
					ID:                p.Post.ID,
					Title:             "reply",
					Class:             "transparent circle",
					Icon:              "reply",
					OnClickActionName: p.OnClickReplyActionName,
					Disabled:          p.ButtonsDisabled,
				},
			)
		}),

		app.If(p.LoggedUserNickname == p.Post.Nickname, func() app.UI {
			return app.Div().Body(
				app.B().Title("reaction count").Text(p.Post.ReactionCount).Class("left-padding"),

				&atoms.Button{
					ID:                p.Post.ID,
					Title:             "delete this post",
					Class:             "transparent circle",
					Icon:              "delete",
					OnClickActionName: p.OnClickDeleteActionName,
					Disabled:          p.ButtonsDisabled,
				},
			)
		}).ElseIf(p.Post.Nickname == "system", func() app.UI {
			return app.Div()
		}).Else(func() app.UI {
			return app.Div().Body(
				app.B().Title("reaction count").Text(p.Post.ReactionCount).Class("left-padding"),

				&atoms.Button{
					ID:                p.Post.ID,
					Title:             "increase the reaction count",
					Class:             "transparent circle",
					Icon:              "ac_unit",
					OnClickActionName: p.OnClickStarActionName,
					Disabled:          p.ButtonsDisabled,
					Attr:              map[string]string{"touch-action": "none"},
				},
			)
		}),
	)
}
