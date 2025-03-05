package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type FlowHeader struct {
	app.Compo

	SingleUser models.User

	SinglePostID string
	Hashtag      string

	ButtonsDisabled bool
	RefreshClicked  bool
}

func (h *FlowHeader) Render() app.UI {
	var heading = func() string {
		if h.Hashtag != "" && len(h.Hashtag) < 20 {
			return "hashtag #" + h.Hashtag
		}
		if h.Hashtag != "" && len(h.Hashtag) >= 20 {
			return "hashtag"
		}
		if h.SinglePostID != "" {
			return "single post and replies"
		}

		return "flow"
	}

	return app.Div().Class("row").Body(
		app.Div().Class("max padding").Body(
			app.If(h.SingleUser.Nickname != "", func() app.UI {
				return app.H5().Body(
					app.Text(h.SingleUser.Nickname+"'s flow"),

					app.If(h.SingleUser.Private, func() app.UI {
						return app.Span().Class("bold").Body(
							app.I().Text("lock"),
						)
					}),
				)
			}).Else(func() app.UI {
				return app.H5().Text(heading())
			}),
		),

		app.Div().Class("small-padding").Body(
			&atoms.Button{
				ID:                "refresh-button",
				Title:             "refresh flow [R]",
				Class:             "primary-container white-text bold thicc",
				Icon:              "refresh",
				Text:              "Refresh",
				OnClickActionName: "refresh",
				Disabled:          h.ButtonsDisabled,
				ShowProgress:      h.RefreshClicked,
			},
		),
	)
}
