package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"
)

type PostFeed struct {
	app.Compo

	SortedPosts []models.Post

	LoggedUserNickname string

	Posts map[string]models.Post
	Users map[string]models.User

	ButtonsDisabled bool
	LoaderShowImage bool

	OnClickDeleteButton app.EventHandler
	OnClickImage        app.EventHandler
	OnClickLink         app.EventHandler
	OnClickReply        app.EventHandler
	OnClickStar         app.EventHandler
	OnClickUser         app.EventHandler
	OnMouseEnter        app.EventHandler
	OnMouseLeave        app.EventHandler
}

func (p *PostFeed) Render() app.UI {
	return app.Div().Class("post-feed").Body(
		app.Range(p.SortedPosts).Slice(func(idx int) app.UI {
			post := p.SortedPosts[idx]

			return app.Div().Class("post").Body(
				&molecules.PostHeader{
					PostAuthor:      post.Nickname,
					PostAvatarURL:   p.Users[post.Nickname].AvatarURL,
					PostID:          post.ID,
					ButtonsDisabled: p.ButtonsDisabled,
					OnClickLink:     p.OnClickLink,
					OnClickUser:     p.OnClickUser,
					OnMouseEnter:    p.OnMouseEnter,
					OnMouseLeave:    p.OnMouseLeave,
				},
				&molecules.PostBody{
					Post:            post,
					PostOriginal:    p.Posts[post.ReplyToID],
					ReplyToID:       post.ReplyToID,
					ButtonDisabled:  p.ButtonsDisabled,
					LoaderShowImage: p.LoaderShowImage,
					OnClickImage:    p.OnClickImage,
					OnClickLink:     p.OnClickLink,
				},
				&molecules.PostFooter{
					Post:                post,
					ButtonsDisabled:     p.ButtonsDisabled,
					LoggedUserNickname:  p.LoggedUserNickname,
					OnClickDeleteButton: p.OnClickDeleteButton,
					OnClickStar:         p.OnClickStar,
					OnClickReply:        p.OnClickReply,
				},
			)
		}),
	)
}
