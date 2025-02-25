package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type SingleUserSummary struct {
	app.Compo

	LoggedUser models.User
	SingleUser models.User

	OnClickFollowActionName string

	ButtonsDisabled bool
}

func (s *SingleUserSummary) Render() app.UI {
	return app.Div().Body(
		app.If(s.SingleUser.Nickname != "", func() app.UI {
			return app.Div().Body(
				&atoms.Image{
					Class:  "center",
					Src:    s.SingleUser.AvatarURL,
					Styles: map[string]string{"max-width": "15rem", "border-radius": "50%"},
				},

				app.Div().Class("row top-padding").Body(
					app.Article().Class("max thicc border").Style("word-break", "break-word").Style("hyphens", "auto").Text(s.SingleUser.About),

					app.If(s.LoggedUser.FlowList[s.SingleUser.Nickname], func() app.UI {
						return &atoms.Button{
							ID:                s.SingleUser.Nickname,
							Class:             "grey10 white-text thicc",
							Icon:              "close",
							Text:              "Unfollow",
							OnClickActionName: s.OnClickFollowActionName,
							Disabled:          s.ButtonsDisabled || s.SingleUser.Nickname == s.LoggedUser.Nickname,
						}
					}).ElseIf(s.SingleUser.Private || s.SingleUser.Options["private"], func() app.UI {
						return &atoms.Button{
							ID:       s.SingleUser.Nickname,
							Class:    "yellow10 white-text thicc",
							Icon:     "drafts",
							Text:     "Ask",
							OnClick:  nil,
							Disabled: s.ButtonsDisabled || s.SingleUser.Nickname == s.LoggedUser.Nickname,
						}
					}).Else(func() app.UI {
						return &atoms.Button{
							ID:                s.SingleUser.Nickname,
							Class:             "deep-orange7 white-text thicc",
							Icon:              "add",
							Text:              "Follow",
							OnClickActionName: s.OnClickFollowActionName,
							Disabled:          s.ButtonsDisabled || s.SingleUser.Nickname == s.LoggedUser.Nickname,
						}
					}),
				),
			)
		}),
	)
}
