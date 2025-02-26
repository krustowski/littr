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

	OnClickLinkActionName  string
	OnClickUserActionName  string
	OnMouseEnterActionName string
	OnMouseLeaveActionName string
}

func (p *PostHeader) Render() app.UI {
	// post header (author avatar + name + link button)
	return app.Div().Class("row top-padding bottom-padding").Body(
		&atoms.Image{
			ID:                p.PostAuthor,
			Title:             "user's avatar",
			Class:             "responsive max left",
			Src:               p.PostAvatarURL,
			Styles:            map[string]string{"max-width": "60px", "border-radius": "50%"},
			OnClickActionName: p.OnClickUserActionName,
		},

		&atoms.UserNickname{
			SpanID:                 "user-flow-link-" + p.PostID,
			Title:                  "user's flow link",
			Class:                  "large-text bold deep-orange-text",
			Nickname:               p.PostAuthor,
			OnClickActionName:      p.OnClickLinkActionName,
			OnMouseEnterActionName: p.OnMouseEnterActionName,
			OnMouseLeaveActionName: p.OnMouseLeaveActionName,
		},

		&atoms.Button{
			ID:                p.PostID,
			Title:             "link to this post",
			Class:             "transparent circle",
			Icon:              "link",
			OnClickActionName: p.OnClickLinkActionName,
			Disabled:          p.ButtonsDisabled,
		},
	)
}
