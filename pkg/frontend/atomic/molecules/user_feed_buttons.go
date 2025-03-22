package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type UserFeedButtons struct {
	app.Compo

	LoggedUserNickname string
	User               models.User

	IsInFlow    bool
	IsPrivate   bool
	IsRequested bool
	IsShaded    bool

	ButtonsDisabled bool

	OnClickAskActionName      string
	OnClickShadeActionName    string
	OnClickCancelActionName   string
	OnClickFollowActionName   string
	OnClickUnfollowActionName string
}

func (b *UserFeedButtons) Render() app.UI {
	return app.Div().Class("").Body(
		app.If(b.User.Nickname == "system", func() app.UI {
			return nil
		}),

		app.If(b.User.Nickname == b.LoggedUserNickname, func() app.UI {
			// When the b.User.is followed.
			return app.Div().Class("row").Body(
				&atoms.Button{
					Class:    "max responsive shrink grey white-text thicc",
					Disabled: true,
					Text:     "That's you",
					Icon:     "cruelty_free",
				},

				&atoms.Button{
					Class:    "max responsive shrink grey10 white-border white-text bold thicc",
					Disabled: true,
					Icon:     "close",
					Text:     "That's you",
				},
			)
		}).ElseIf(b.IsShaded, func() app.UI {
			// When the b.User.is shaded, all actions are blocked by default.
			return app.Div().Class("row").Body(
				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey white-text thicc",
					Disabled:          !b.IsShaded,
					Text:              "Unshade",
					Icon:              "cruelty_free",
					OnClickActionName: b.OnClickShadeActionName,
				},

				&atoms.Button{
					ID:       b.User.Nickname,
					Class:    "max responsive shrink grey10 white-border white-text bold thicc",
					Disabled: b.IsShaded,
					Icon:     "close",
					Text:     "Shaded",
				},
			)
		}).ElseIf(b.IsInFlow, func() app.UI {
			// When the b.User.is followed.
			return app.Div().Class("row").Body(
				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey black-text thicc",
					Disabled:          b.IsShaded,
					Text:              "Shade",
					Icon:              "block",
					OnClickActionName: b.OnClickShadeActionName,
				},

				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey10 white-border white-text bold thicc",
					Disabled:          b.ButtonsDisabled,
					Icon:              "close",
					Text:              "Unfollow",
					OnClickActionName: b.OnClickUnfollowActionName,
				},
			)
		}).ElseIf(!b.IsInFlow && b.IsPrivate && !b.IsRequested, func() app.UI {
			return app.Div().Class("row").Body(
				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey black-text thicc",
					Disabled:          b.IsShaded,
					Text:              "Shade",
					Icon:              "block",
					OnClickActionName: b.OnClickShadeActionName,
				},

				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink blue10 white-border white-text bold thicc",
					Disabled:          b.ButtonsDisabled,
					Icon:              "drafts",
					Text:              "Ask to follow",
					OnClickActionName: b.OnClickAskActionName,
				},
			)
		}).ElseIf(!b.IsInFlow && b.IsRequested && b.IsPrivate, func() app.UI {
			return app.Div().Class("row").Body(
				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey black-text thicc",
					Disabled:          b.IsShaded,
					Text:              "Shade",
					Icon:              "block",
					OnClickActionName: b.OnClickShadeActionName,
				},

				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey10 white-border white-text bold thicc",
					Disabled:          b.ButtonsDisabled,
					Icon:              "close",
					Text:              "Cancel the follow request",
					OnClickActionName: b.OnClickCancelActionName,
				},
			)
		}).Else(func() app.UI {
			return app.Div().Class("row").Body(
				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink grey black-text thicc",
					Disabled:          b.IsShaded,
					Text:              "Shade",
					Icon:              "block",
					OnClickActionName: b.OnClickShadeActionName,
				},

				&atoms.Button{
					ID:                b.User.Nickname,
					Class:             "max responsive shrink primary-container white-border white-text bold thicc",
					Disabled:          b.ButtonsDisabled,
					Icon:              "add",
					Text:              "Follow",
					OnClickActionName: b.OnClickFollowActionName,
				},
			)
		}),
	)
}
