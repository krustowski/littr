package frontend

import (
	"encoding/json"
	"log"
	"strconv"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type PollsPage struct {
	app.Compo
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

type pollsContent struct {
	app.Compo

	loggedUser string
	user       models.User

	loaderShow bool

	toastShow bool
	toastText string

	polls map[string]models.Poll
}

func (c *pollsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *pollsContent) OnNav(ctx app.Context) {
	// show loader
	c.loaderShow = true

	ctx.Async(func() {
		pollsRaw := struct {
			Polls map[string]models.Poll `json:"polls"`
		}{}

		if byteData, _ := litterAPI("GET", "/api/polls", nil); byteData != nil {
			err := json.Unmarshal(*byteData, &pollsRaw)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else {
			log.Println("cannot fetch polls list")
			return
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.polls = pollsRaw.Polls
			//c.sortedPosts = posts

			c.loaderShow = false
			log.Println("dispatch ends")
		})
	})
	return
}

func (c *pollsContent) onClickPollOption(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	option := ctx.JSSrc().Get("text").String()
	ctx.NewActionWithValue("vote", []string{key, option})
}

func (c *pollsContent) handleVote(ctx app.Context, a app.Action) {
	keys, ok := a.Value.([]string)
	if !ok {
		return
	}

	key := keys[0]
	//option := keys[1]

	ctx.Async(func() {
		var toastText string = ""

		interactedPoll := c.polls[key]

		if _, ok := litterAPI("PUT", "/api/polls", interactedPoll); !ok {
			toastText = "backend error: cannot update a poll"
		}

		ctx.Dispatch(func(ctx app.Context) {
			//delete(c.polls, key)

			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
	})
}

func (c *pollsContent) onClickDelete(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("delete", key)
}

func (c *pollsContent) handleDelete(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		var toastText string = ""

		interactedPoll := c.polls[key]

		if _, ok := litterAPI("DELETE", "/api/polls", interactedPoll); !ok {
			toastText = "backend error: cannot delete a poll"
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.polls, key)

			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
	})
}

func (c *pollsContent) OnMount(ctx app.Context) {
	var enUser string
	var user models.User
	var toastText string = ""

	ctx.Handle("vote", c.handleVote)
	ctx.Handle("delete", c.handleDelete)

	ctx.LocalStorage().Get("user", &enUser)
	// decode, decrypt and unmarshal the local storage string
	if err := prepare(enUser, &user); err != nil {
		toastText = "frontend decoding/decryption failed: " + err.Error()
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.user = user
		c.loggedUser = user.Nickname
		c.toastText = toastText
		c.toastShow = (toastText != "")
	})
	return
}

// contains checks if a string is present in a slice
// https://freshman.tech/snippets/go/check-if-slice-contains-element/
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func (c *pollsContent) Render() app.UI {
	loaderActiveClass := ""
	if c.loaderShow {
		loaderActiveClass = " active"
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
					app.Th().Class("align-left").Text("poll question, options, timestamp"),
				),
			),
			app.TBody().Body(
				app.Range(c.polls).Map(func(key string) app.UI {
					poll := c.polls[key]

					userVoted := contains(poll.Voted, c.user.Nickname)

					optionOneShare := 0
					optionTwoShare := 0
					optionThreeShare := 0

					pollCounterSum := 0
					pollCounterSum = poll.OptionOne.Counter + poll.OptionTwo.Counter
					if poll.OptionThree.Content != "" {
						pollCounterSum += poll.OptionThree.Counter
					}

					// at least one vote has to be already recorded to show the progresses
					if pollCounterSum > 0 {
						optionOneShare = poll.OptionOne.Counter * 100 / pollCounterSum
						optionTwoShare = poll.OptionTwo.Counter * 100 / pollCounterSum
						optionThreeShare = poll.OptionThree.Counter * 100 / pollCounterSum
					}

					return app.Tr().Body(
						app.Td().Class("align-left").Body(
							app.P().Body(
								app.Text("Q: "),
								app.B().Text(poll.Question).Class("deep-orange-text space"),
							),
							app.Div().Class("space"),

							// show buttons to vote
							app.If(!userVoted && poll.Author != c.user.Nickname,
								app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionOne.Content),
								app.Div().Class("space"),
								app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionTwo.Content),
								app.Div().Class("space"),
								app.If(poll.OptionThree.Content != "",
									app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionThree.Content),
									app.Div().Class("space"),
								),

							// show results instead
							).ElseIf(userVoted || poll.Author == c.user.Nickname,
								// voted option I
								app.Div().Class("medium-space border").Body(
									app.P().Class("middle right-align bold padding").Body(
										app.Text(poll.OptionOne.Content),
									),
									app.Div().Class("progress left deep-orange large").
										Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionOneShare)+"% 100%, "+strconv.Itoa(optionOneShare)+"% 0%);").OnClick(c.onClickPollOption).DataSet("option", poll.OptionOne.Content),
								),
								app.Div().Class("space"),

								// voted option II
								app.Div().Class("medium-space border").Body(
									app.P().Class("middle right-align bold padding").Body(
										app.Text(poll.OptionTwo.Content),
									),
									app.Div().Class("progress left deep-orange").
										Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionTwoShare)+"% 100%, "+strconv.Itoa(optionTwoShare)+"% 0%);").OnClick(c.onClickPollOption).DataSet("option", poll.OptionTwo.Content),
								),
								app.Div().Class("space"),

								// voted option III
								app.If(poll.OptionThree.Content != "",
									app.Div().Class("medium-space border").Body(
										app.P().Class("middle bold right-align padding").Text(poll.OptionThree.Content),
										app.Div().Class("progress left deep-orange").
											Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionThreeShare)+"% 100%, "+strconv.Itoa(optionThreeShare)+"% 0%);").OnClick(c.onClickPollOption).DataSet("option", poll.OptionThree.Content),
									),
									app.Div().Class("space"),
								),
							),

							// bottom row of the poll
							app.Div().Class("row").Body(
								app.Div().Class("max").Body(
									app.Text(poll.Timestamp.Format("Jan 02, 2006; 15:04:05")),
								),
								app.If(poll.Author == c.user.Nickname,
									app.B().Text(len(poll.Voted)),
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickDelete).Body(
										app.I().Text("delete"),
									),
								).Else(
									app.B().Text(len(poll.Voted)),
									app.Button().ID(key).Class("transparent circle").OnClick(nil).Body(
										app.I().Text("star"),
									),
								),
							),
						),
					)
				}),
			),
		),
		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.Div().Class("loader center large deep-orange"+loaderActiveClass),
		),
	)
}
