package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type PollBody struct {
	app.Compo

	RenderProps struct {
		PollTimestamp    string
		UserVoted        bool
		OptionOneShare   int64
		OptionTwoShare   int64
		OptionThreeShare int64
	}

	Poll       models.Poll
	LoggedUser models.User

	OnClickOptionOneActionName   string
	OnClickOptionTwoActionName   string
	OnClickOptionThreeActionName string

	ButtonDisabled  bool
	LoaderShowImage bool
}

func (p *PollBody) OnMount(ctx app.Context) {}

func (p *PollBody) Render() app.UI {
	return app.Div().Body(
		app.If(!p.RenderProps.UserVoted && p.Poll.Author != p.LoggedUser.Nickname, func() app.UI {
			return app.Div().Body(
				&atoms.Button{
					ID:                p.Poll.ID,
					Name:              p.Poll.OptionOne.Content,
					Title:             "option one",
					Class:             "blue10 bold white-text responsive thicc",
					Text:              p.Poll.OptionOne.Content,
					OnClickActionName: p.OnClickOptionOneActionName,
					Disabled:          p.ButtonDisabled,
					DataSet:           map[string]string{"option": p.Poll.OptionOne.Content},
				},

				app.Div().Class("space"),

				&atoms.Button{
					ID:                p.Poll.ID,
					Name:              p.Poll.OptionTwo.Content,
					Title:             "option two",
					Class:             "blue10 bold white-text responsive thicc",
					Text:              p.Poll.OptionTwo.Content,
					OnClickActionName: p.OnClickOptionTwoActionName,
					Disabled:          p.ButtonDisabled,
					DataSet:           map[string]string{"option": p.Poll.OptionTwo.Content},
				},

				app.Div().Class("space"),

				app.If(p.Poll.OptionThree.Content != "", func() app.UI {
					return app.Div().Body(
						&atoms.Button{
							ID:                p.Poll.ID,
							Name:              p.Poll.OptionThree.Content,
							Title:             "option three",
							Class:             "blue10 bold white-text responsive thicc",
							Text:              p.Poll.OptionThree.Content,
							OnClickActionName: p.OnClickOptionThreeActionName,
							Disabled:          p.ButtonDisabled,
							DataSet:           map[string]string{"option": p.Poll.OptionThree.Content},
						},
						app.Div().Class("space"),
					)
				}),
			)
		}).ElseIf(p.RenderProps.UserVoted || p.Poll.Author == p.LoggedUser.Nickname, func() app.UI {
			return app.Div().Body(
				// Option one results.
				&atoms.PollResult{
					OptionShare:  p.RenderProps.OptionOneShare,
					Option:       p.Poll.OptionOne,
					OrangelLevel: 3,
				},
				app.Div().Class("medium-space"),

				// Option two results.
				&atoms.PollResult{
					OptionShare:  p.RenderProps.OptionTwoShare,
					Option:       p.Poll.OptionTwo,
					OrangelLevel: 5,
				},
				app.Div().Class("medium-space"),

				// Option three results (if present).
				app.If(p.Poll.OptionThree.Content != "", func() app.UI {
					return app.Div().Body(
						&atoms.PollResult{
							OptionShare:  p.RenderProps.OptionThreeShare,
							Option:       p.Poll.OptionThree,
							OrangelLevel: 9,
						},
						app.Div().Class("medium-space"),
					)
				}),
			)
		}),
	)
}
