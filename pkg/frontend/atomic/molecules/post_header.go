package molecules

import (
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type PostHeader struct {
	app.Compo

	PostAuthor    string
	PostAvatarURL string
	PostID        string

	ButtonsDisabled bool

	OnClickLink  app.EventHandler
	OnClickUser  app.EventHandler
	OnMouseEnter app.EventHandler
	OnMouseLeave app.EventHandler
}

func (p *PostHeader) Render() app.UI {
	// post header (author avatar + name + link button)
	return app.Div().Class("row top-padding bottom-padding").Body(
		&atoms.Image{
			ID:      p.PostAuthor,
			Title:   "user's avatar",
			Class:   "responsive max left",
			Src:     p.PostAvatarURL,
			Width:   "60px",
			Radius:  "50%",
			OnClick: p.OnClickUser,
		},

		app.P().Class("max").Body(
			app.A().Title("user's flow link").Class("bold deep-orange-text").OnClick(p.OnClickUser).ID(p.PostAuthor).Body(
				app.Span().ID("user-flow-link").Class("large-text bold deep-orange-text").Text(p.PostAuthor).OnMouseEnter(p.OnMouseEnter).OnMouseLeave(p.OnMouseLeave),
			),
			//app.B().Text(post.Nickname).Class("deep-orange-text"),
		),

		&atoms.Button{
			ID:       p.PostID,
			Title:    "link to this post",
			Class:    "transparent circle",
			Icon:     "link",
			OnClick:  p.OnClickLink,
			Disabled: p.ButtonsDisabled,
		},
	)
}
