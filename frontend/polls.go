package frontend

import (
	"strconv"

	"litter-go/config"
	"litter-go/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type PollsPage struct {
	app.Compo
}

type pollsContent struct {
	app.Compo

	polls []models.Poll

	toastShow bool
	toastText string
}

func (p *PollsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("polls / littr")
}

func (p *PollsPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&pollsContent{},
	)
}

func (c *pollsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *pollsContent) Render() app.UI {
	c.polls = []models.Poll{
		{
			Question: "wtf???",
			OptionOne: models.PollOption{
				Content: "lmao",
				Counter: 5,
			},
			OptionTwo: models.PollOption{
				Content: "hm",
				Counter: 2,
			},
			OptionThree: models.PollOption{
				Content: "idk",
				Counter: 10,
			},
		},
	}

	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr polls").Style("padding-top", config.HeaderTopPadding),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Table().Class("border left-align").Body(
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("align-left").Text("nickname, content, timestamp"),
				),
			),
			app.TBody().Body(
				app.Range(c.polls).Slice(func(i int) app.UI {
					poll := c.polls[i]

					pollCounterSum := 0
					pollCounterSum = poll.OptionOne.Counter + poll.OptionTwo.Counter
					if poll.OptionThree.Content != "" {
						pollCounterSum += poll.OptionThree.Counter
					}

					optionOneShare := poll.OptionOne.Counter * 100 / pollCounterSum
					optionTwoShare := poll.OptionTwo.Counter * 100 / pollCounterSum
					optionThreeShare := poll.OptionThree.Counter * 100 / pollCounterSum

					return app.Tr().Body(
						app.Td().Class("align-left").Body(
							app.B().Text(poll.Question).Class("deep-orange-text"),
							app.Div().Class("space"),

							app.Text(poll.OptionOne.Content),
							app.Div().Class("small-space border").Body(
								app.Div().Class("progress left deep-orange").
									Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionOneShare)+"% 100%, "+strconv.Itoa(optionOneShare)+"% 0%);"),
							),
							app.Div().Class("space"),

							app.Text(poll.OptionTwo.Content),
							app.Div().Class("small-space border").Body(
								app.Div().Class("progress left deep-orange").
									Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionTwoShare)+"% 100%, "+strconv.Itoa(optionTwoShare)+"% 0%);"),
							),
							app.Div().Class("space"),

							app.If(poll.OptionThree.Content != "",
								app.Text(poll.OptionThree.Content),
								app.Div().Class("small-space border").Body(
									app.Div().Class("progress left deep-orange").
										Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionThreeShare)+"% 100%, "+strconv.Itoa(optionThreeShare)+"% 0%);"),
								),
								app.Div().Class("space"),
							),

							app.Text(poll.Timestamp.Format("Jan 02, 2006; 15:04:05")),
						),
					)
				}),
			),
		),
	)
}
