package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type PollFooter struct {
	app.Compo

	Poll models.Poll

	LoggedUserNickname string
	PollTimestamp      string

	ButtonsDisabled bool

	OnClickDeleteActionName string
}

func (p *PollFooter) Render() app.UI {
	// bottom row of the poll
	return app.Div().Class("row").Body(
		app.Div().Class("max").Body(
			//app.Text(poll.Timestamp.Format("Jan 02, 2006; 15:04:05")),
			app.Text(p.PollTimestamp),
		),
		app.If(p.Poll.Author == p.LoggedUserNickname, func() app.UI {
			return app.Div().Body(
				app.B().Title("vote count").Text(len(p.Poll.Voted)),

				&atoms.Button{
					ID:                p.Poll.ID,
					Title:             "delete this poll",
					Class:             "transparent circle",
					Icon:              "delete",
					OnClickActionName: p.OnClickDeleteActionName,
					Disabled:          p.ButtonsDisabled,
				},
			)
		}).Else(func() app.UI {
			return app.Div().Body(
				app.B().Title("vote count").Text(len(p.Poll.Voted)),

				&atoms.Button{
					ID:                p.Poll.ID,
					Title:             "voting enabled",
					Class:             "transparent circle",
					Icon:              "how_to_vote",
					OnClickActionName: "",
					Disabled:          true,
				},
			)
		}),
	)
}
