package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"
)

type SingleUserProfile struct {
	app.Compo

	LoggedUser models.User
	SingleUser models.User

	OnClickAskActionName      string
	OnClickShadeActionName    string
	OnClickCancelActionName   string
	OnClickFollowActionName   string
	OnClickUnfollowActionName string

	ButtonsDisabled bool
}

func (s *SingleUserProfile) Render() app.UI {
	var isInFlow, isRequested, isShaded, found bool

	if s.LoggedUser.FlowList != nil {
		if isInFlow, found = s.LoggedUser.FlowList[s.SingleUser.Nickname]; found && isInFlow {
			isInFlow = true
		}
	}

	if s.LoggedUser.ShadeList != nil {
		if isShaded, found = s.LoggedUser.ShadeList[s.SingleUser.Nickname]; found && isShaded {
			isShaded = true
		}
	}

	if s.SingleUser.RequestList != nil {
		if isRequested, found = s.SingleUser.RequestList[s.LoggedUser.Nickname]; !found {
			isRequested = false
		}
	}

	return app.Div().Body(
		app.If(s.SingleUser.Nickname != "", func() app.UI {
			return app.Div().Body(
				&atoms.Image{
					Class:  "center",
					Src:    s.SingleUser.AvatarURL,
					Styles: map[string]string{"max-width": "15rem", "border-radius": "50%"},
				},

				&molecules.TextBox{
					Class: "row border thicc",
					Icon:  "",
					Text:  s.SingleUser.About,
				},
				app.Div().Class("space"),

				&molecules.UserFeedButtons{
					LoggedUserNickname: s.LoggedUser.Nickname,
					User:               s.SingleUser,

					IsInFlow:        isInFlow,
					IsPrivate:       s.SingleUser.Private,
					IsRequested:     isRequested,
					IsShaded:        isShaded,
					ButtonsDisabled: s.ButtonsDisabled,

					OnClickAskActionName:      s.OnClickAskActionName,
					OnClickShadeActionName:    s.OnClickShadeActionName,
					OnClickCancelActionName:   s.OnClickCancelActionName,
					OnClickFollowActionName:   s.OnClickFollowActionName,
					OnClickUnfollowActionName: s.OnClickUnfollowActionName,
				},
			)
		}),
	)
}
