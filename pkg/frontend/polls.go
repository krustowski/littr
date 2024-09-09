package frontend

import (
	"encoding/json"
	"log"
	"sort"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/models"

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
		&header{},
		&footer{},
		&pollsContent{},
	)
}

type pollsContent struct {
	app.Compo

	eventListener func()

	loggedUser string
	user       models.User

	loaderShow bool

	toastShow bool
	toastText string
	toastType string

	paginationEnd bool
	pagination    int
	pageNo        int

	pollKey                    string
	interactedPollKey          string
	deleteModalButtonsDisabled bool
	deletePollModalShow        bool

	polls map[string]models.Poll

	pollsButtonDisabled bool

	keyDownEventListener func()
}

func (c *pollsContent) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = ""
		c.toastShow = false
		c.pollsButtonDisabled = false
		c.deletePollModalShow = false
	})
}

func (c *pollsContent) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}
}

func (c *pollsContent) onClickDismiss(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *pollsContent) OnNav(ctx app.Context) {
	// show loader
	c.loaderShow = true
	var toastText string
	var toastType string

	ctx.Async(func() {
		pollsRaw := struct {
			Polls map[string]models.Poll `json:"polls"`
			Code  int                    `json:"code"`
			Users map[string]models.User `json:"users"`
			Key   string                 `json:"key"`
		}{}

		if byteData, _ := littrAPI("GET", "/api/v1/polls", nil, "", 0); byteData != nil {
			err := json.Unmarshal(*byteData, &pollsRaw)
			if err != nil {
				log.Println(err.Error())

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = err.Error()
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			toastText = "cannot fetch polls list"
			log.Println(toastText)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if pollsRaw.Code == 401 {
			toastText = "please log-in again"

			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if len(pollsRaw.Polls) < 1 {
			toastText = "there is nothing here yet, be the first to create a poll!"
			toastType = "info"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastType = toastType
				c.toastShow = (toastText != "")
				c.loaderShow = false
			})
			return
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.user = pollsRaw.Users[pollsRaw.Key]
			c.loggedUser = c.user.Nickname

			c.pagination = 10
			c.pageNo = 1

			c.polls = pollsRaw.Polls

			c.pollsButtonDisabled = false
			c.loaderShow = false
		})
	})
	return
}

func (c *pollsContent) onClickDeleteButton(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPollKey = key
		c.deleteModalButtonsDisabled = false
		c.deletePollModalShow = true
	})
}

func (c *pollsContent) onClickDelete(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("delete", c.interactedPollKey)

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteModalButtonsDisabled = true
	})
}

func (c *pollsContent) handleDelete(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	if key != c.interactedPollKey {
		return
	}

	ctx.Async(func() {
		var toastText string = ""

		//key := c.pollKey
		interactedPoll := c.polls[key]

		if interactedPoll.Author != c.user.Nickname {
			toastText = "you only can delete your own polls!"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.deletePollModalShow = false
				c.deleteModalButtonsDisabled = false
			})
			return
		}

		if _, ok := littrAPI("DELETE", "/api/v1/polls/"+interactedPoll.ID, interactedPoll, c.user.Nickname, c.pageNo); !ok {
			toastText = "backend error: cannot delete a poll"
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.polls, key)

			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.deletePollModalShow = false
			c.deleteModalButtonsDisabled = false
		})
	})
}

func (c *pollsContent) onScroll(ctx app.Context, e app.Event) {
	ctx.NewAction("scroll")
}

func (c *pollsContent) handleScroll(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		if bottom-height < 0 && !c.paginationEnd {
			ctx.Dispatch(func(ctx app.Context) {
				c.pageNo++
				log.Println("new content page request fired")
			})
			return
		}
	})
}

func (c *pollsContent) onClickPollOption(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	option := ctx.JSSrc().Get("name").String()

	ctx.NewActionWithValue("vote", []string{key, option})

	c.pollsButtonDisabled = true
}

func (c *pollsContent) handleVote(ctx app.Context, a app.Action) {
	keys, ok := a.Value.([]string)
	if !ok {
		return
	}

	key := keys[0]
	option := keys[1]

	poll := c.polls[key]
	toastText := ""

	poll.Voted = append(poll.Voted, c.user.Nickname)

	// check where to vote
	options := []string{
		poll.OptionOne.Content,
		poll.OptionTwo.Content,
		poll.OptionThree.Content,
	}

	// use the vote
	if found := contains(options, option); found {
		switch option {
		case poll.OptionOne.Content:
			poll.OptionOne.Counter++
			break

		case poll.OptionTwo.Content:
			poll.OptionTwo.Counter++
			break

		case poll.OptionThree.Content:
			poll.OptionThree.Counter++
			break
		}
	} else {
		toastText = "option not associated to the poll well"
	}

	ctx.Async(func() {
		//var toastText string

		if _, ok := littrAPI("PUT", "/api/v1/polls/"+poll.ID, poll, c.user.Nickname, 0); !ok {
			toastText = "backend error: cannot update a poll"
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.polls[key] = poll

			c.pollsButtonDisabled = false
			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
	})
}

func (c *pollsContent) OnMount(ctx app.Context) {
	ctx.Handle("vote", c.handleVote)
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("dismiss", c.handleDismiss)

	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1

	c.eventListener = app.Window().AddEventListener("scroll", c.onScroll)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
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
	toastColor := ""

	switch c.toastType {
	case "success":
		toastColor = "green10"
		break

	case "info":
		toastColor = "blue10"
		break

	default:
		toastColor = "red10"
	}

	var sortedPolls []models.Poll

	for _, sortedPoll := range c.polls {
		sortedPolls = append(sortedPolls, sortedPoll)
	}

	// order polls by timestamp DESC
	sort.SliceStable(sortedPolls, func(i, j int) bool {
		return sortedPolls[i].Timestamp.After(sortedPolls[j].Timestamp)
	})

	// prepare polls according to the actual pagination and pageNo
	pagedPolls := []models.Poll{}

	end := len(sortedPolls)
	start := 0

	stop := func(c *pollsContent) int {
		var pos int

		if c.pagination > 0 {
			// (c.pageNo - 1) * c.pagination + c.pagination
			pos = c.pageNo * c.pagination
		}

		if pos > end {
			// kill the eventListener (observers scrolling)
			c.eventListener()
			c.paginationEnd = true

			return (end)
		}

		if pos < 0 {
			return 0
		}

		return pos
	}(c)

	if end > 0 && stop > 0 {
		pagedPolls = sortedPolls[start:stop]
	}

	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("polls"),
				//app.P().Text("brace yourself"),
			),
		),
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.onClickDismiss).Body(
			app.If(c.toastText != "",
				app.Div().ID("snackbar").Class("snackbar "+toastColor+" white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// poll deletion modal
		app.If(c.deletePollModalShow,
			app.Dialog().ID("delete-modal").Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("poll deletion"),
				),

				app.Div().Class("space"),
				app.Article().Class("row").Body(
					app.I().Text("warning").Class("amber-text"),
					app.P().Class("max").Body(
						app.Span().Text("are you sure you want to delete your poll?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max border red10 white-text").Style("border-radius", "8px").OnClick(c.onClickDelete).Disabled(c.deleteModalButtonsDisabled).Body(
						app.If(c.deleteModalButtonsDisabled,
							app.Progress().Class("circle white-border small"),
						),
						app.Text("yeah"),
					),
					app.Button().Class("max border black white-text").Style("border-radius", "8px").Text("nope").OnClick(c.onClickDismiss).Disabled(c.deleteModalButtonsDisabled),
				),
			),
		),

		app.Table().Class("left-align border").ID("table-poll").Style("padding", "0 0 2em 0").Style("border-spacing", "0.1em").Body(
			app.TBody().Body(
				app.Range(pagedPolls).Slice(func(idx int) app.UI {
					poll := pagedPolls[idx]
					key := poll.ID

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

					var pollTimestamp string

					// use JS toLocaleString() function to reformat the timestamp
					// use negated LocalTimeMode boolean as true! (HELP)
					if !c.user.LocalTimeMode {
						pollLocale := app.Window().
							Get("Date").
							New(poll.Timestamp.Format(time.RFC3339))

						pollTimestamp = pollLocale.Call("toLocaleString", "en-GB").String()
					} else {
						pollTimestamp = poll.Timestamp.Format("Jan 02, 2006 / 15:04:05")
					}

					return app.Tr().Body(
						app.Td().Attr("data-timestamp", poll.Timestamp.UnixNano()).Class("align-left").Body(
							app.Div().Class("row top-padding").Body(
								app.P().Body(
									app.Span().Title("question").Text("Q: "),
									app.Span().Text(poll.Question).Class("deep-orange-text space bold"),
								),
							),
							app.Div().Class("space"),

							// show buttons to vote
							app.If(!userVoted && poll.Author != c.user.Nickname,
								app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionOne.Content).DataSet("option", poll.OptionOne.Content).OnClick(c.onClickPollOption).ID(poll.ID).Name(poll.OptionOne.Content).Disabled(c.pollsButtonDisabled).Style("border-radius", "8px"),
								app.Div().Class("space"),
								app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionTwo.Content).DataSet("option", poll.OptionTwo.Content).OnClick(c.onClickPollOption).ID(poll.ID).Name(poll.OptionTwo.Content).Disabled(c.pollsButtonDisabled).Style("border-radius", "8px"),
								app.Div().Class("space"),
								app.If(poll.OptionThree.Content != "",
									app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionThree.Content).DataSet("option", poll.OptionThree.Content).OnClick(c.onClickPollOption).ID(poll.ID).Name(poll.OptionThree.Content).Disabled(c.pollsButtonDisabled).Style("border-radius", "8px"),
									app.Div().Class("space"),
								),

							// show results instead
							).ElseIf(userVoted || poll.Author == c.user.Nickname,

								// voted option I
								app.Div().Class("medium-space border").Body(
									app.Div().Class("bold progress left deep-orange3 medium padding").Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionOneShare)+"% 100%, "+strconv.Itoa(optionOneShare)+"% 0%);"),
									//app.Progress().Value(strconv.Itoa(optionOneShare)).Max(100).Class("deep-orange-text padding medium bold left"),
									//app.Div().Class("progress left light-green"),
									app.Div().Class("middle right-align bold").Body(
										app.Span().Text(poll.OptionOne.Content+" ("+strconv.Itoa(optionOneShare)+"%)"),
									),
								),

								app.Div().Class("medium-space"),

								// voted option II
								app.Div().Class("medium-space border").Body(
									app.Div().Class("bold progress left deep-orange5 medium padding").Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionTwoShare)+"% 100%, "+strconv.Itoa(optionTwoShare)+"% 0%);").Body(),
									//app.Progress().Value(strconv.Itoa(optionTwoShare)).Max(100).Class("deep-orange-text padding medium bold left"),
									app.Div().Class("middle right-align bold").Body(
										app.Span().Text(poll.OptionTwo.Content+" ("+strconv.Itoa(optionTwoShare)+"%)"),
									),
								),

								app.Div().Class("space"),

								// voted option III
								app.If(poll.OptionThree.Content != "",
									app.Div().Class("space"),
									app.Div().Class("medium-space border").Body(
										app.Div().Class("bold progress left deep-orange9 medium padding").Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.Itoa(optionThreeShare)+"% 100%, "+strconv.Itoa(optionThreeShare)+"% 0%);"),
										//app.Progress().Value(strconv.Itoa(optionThreeShare)).Max(100).Class("deep-orange-text deep-orange padding medium bold left"),
										app.Div().Class("middle bold right-align").Body(
											app.Span().Text(poll.OptionThree.Content+" ("+strconv.Itoa(optionThreeShare)+"%)"),
										),
									),

									app.Div().Class("space"),
								),
							),

							// bottom row of the poll
							app.Div().Class("row").Body(
								app.Div().Class("max").Body(
									//app.Text(poll.Timestamp.Format("Jan 02, 2006; 15:04:05")),
									app.Text(pollTimestamp),
								),
								app.If(poll.Author == c.user.Nickname,
									app.B().Title("vote count").Text(len(poll.Voted)),
									app.Button().Title("delete this poll").ID(key).Class("transparent circle").OnClick(c.onClickDeleteButton).Body(
										app.I().Text("delete"),
									),
								).Else(
									app.B().Title("vote count").Text(len(poll.Voted)),
									app.Button().Title("just voting allowed").ID(key).Class("transparent circle").Disabled(true).Body(
										app.I().Text("how_to_vote"),
									),
								),
							),
						),
					)
				}),
			),
		),
		app.Div().ID("page-end-anchor"),
		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.Progress().Class("circle center large deep-orange-border active"),
		),
	)
}
