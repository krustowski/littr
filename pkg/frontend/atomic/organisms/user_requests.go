package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type UserRequests struct {
	app.Compo

	LoggedUser models.User
	Users      map[string]models.User

	OnClickAllowActionName  string
	OnClickCancelActionName string
	OnClickUserActionName   string
	OnMouseEnterActionName  string
	OnMouseLeaveActionName  string

	ButtonsDisabled bool
}

func (u *UserRequests) Render() app.UI {
	return app.Div().Class("post-feed").Body(
		app.Range(u.LoggedUser.RequestList).Map(func(key string) app.UI {
			if !u.LoggedUser.RequestList[key] {
				return nil
			}

			return app.Div().Class("post").Body(
				app.Div().Class("row medium top-padding").Body(
					&atoms.Image{
						ID:     key,
						Class:  "responsive max left thicc",
						Src:    u.Users[key].AvatarURL,
						Styles: map[string]string{"max-width": "60px"},
					},

					&atoms.UserNickname{
						Class:                  "deep-orange-text bold max large-text",
						Nickname:               key,
						SpanID:                 key,
						Title:                  "user's nickname",
						Text:                   key,
						OnClickActionName:      u.OnClickUserActionName,
						OnMouseEnterActionName: u.OnMouseEnterActionName,
						OnMouseLeaveActionName: u.OnMouseLeaveActionName,
					},

					&atoms.Button{
						ID:                key,
						Class:             "max responsive no-padding bold grey10 white-text thicc",
						Icon:              "close",
						Text:              "Cancel",
						OnClickActionName: u.OnClickCancelActionName,
						Disabled:          u.ButtonsDisabled,
					},

					&atoms.Button{
						ID:                key,
						Class:             "max responsive no-padding bold deep-orange7 white-text thicc",
						Icon:              "check",
						Text:              "Allow",
						OnClickActionName: u.OnClickAllowActionName,
						Disabled:          u.ButtonsDisabled,
					},
				),
			)
		}),
	)
}
