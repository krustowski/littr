package polls

import (
	"sort"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) Render() app.UI {
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

	stop := func(c *Content) int {
		var pos int

		if c.pagination > 0 {
			// (c.pageNo - 1) * c.pagination + c.pagination
			pos = c.pageNo * c.pagination
		}

		if pos > end {
			// kill the eventListener (observers scrolling)
			c.scrollEventListener()
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
		app.A().Href(c.toast.TLink).OnClick(c.onClickDismiss).Body(
			app.If(c.toast.TText != "", func() app.UI {
				return app.Div().ID("snackbar").Class("snackbar white-text top active "+common.ToastColor(c.toast.TType)).Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				)
			},
			),
		),

		// poll deletion modal
		app.If(c.deletePollModalShow, func() app.UI {
			return app.Div().Body(
				app.Dialog().ID("delete-modal").Class("grey10 white-text active thicc").Body(
					app.Nav().Class("center-align").Body(
						app.H5().Text("poll deletion"),
					),

					app.Div().Class("space"),

					app.Article().Class("row border amber-border white-text warn thicc").Body(
						app.I().Text("warning").Class("amber-text"),
						app.P().Class("max bold").Body(
							app.Span().Text("Are you sure you want to delete your poll?"),
						),
					),
					app.Div().Class("space"),

					app.Div().Class("row").Body(
						app.Button().Class("max bold black white-text thicc").OnClick(c.onClickDismiss).Disabled(c.deleteModalButtonsDisabled).Body(
							app.Span().Body(
								app.I().Style("padding-right", "5px").Text("close"),
								app.Text("Cancel"),
							),
						),
						app.Button().Class("max bold red10 white-text thicc").OnClick(c.onClickDelete).Disabled(c.deleteModalButtonsDisabled).Body(
							app.If(c.deleteModalButtonsDisabled, func() app.UI {
								return app.Progress().Class("circle white-border small")
							}),
							app.Span().Body(
								app.I().Style("padding-right", "5px").Text("delete"),
								app.Text("Delete"),
							),
						),
					),
				),
			)
		},
		),

		app.Table().Class("left-align border").ID("table-poll").Style("padding", "0 0 2em 0").Style("border-spacing", "0.1em").Body(
			app.TBody().Body(
				app.Range(pagedPolls).Slice(func(idx int) app.UI {
					poll := pagedPolls[idx]
					key := poll.ID

					userVoted := contains(poll.Voted, c.user.Nickname)

					var optionOneShare int64
					var optionTwoShare int64
					var optionThreeShare int64

					var pollCounterSum int64
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
								app.P().Class("max").Body(
									app.Span().Title("question").Text("Q: "),
									app.Span().Text(poll.Question).Class("deep-orange-text space bold"),
								),
								app.Button().ID(key).Title("link to this post (to clipboard)").Class("transparent circle").OnClick(c.onClickLink).Disabled(c.pollsButtonDisabled).Body(
									app.I().Text("link"),
								),
							),
							app.Div().Class("space"),

							// show buttons to vote
							app.If(!userVoted && poll.Author != c.user.Nickname, func() app.UI {
								return app.Div().Body(
									app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionOne.Content).DataSet("option", poll.OptionOne.Content).OnClick(c.onClickPollOption).ID(poll.ID).Name(poll.OptionOne.Content).Disabled(c.pollsButtonDisabled).Style("border-radius", "8px"),
									app.Div().Class("space"),
									app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionTwo.Content).DataSet("option", poll.OptionTwo.Content).OnClick(c.onClickPollOption).ID(poll.ID).Name(poll.OptionTwo.Content).Disabled(c.pollsButtonDisabled).Style("border-radius", "8px"),
									app.Div().Class("space"),
									app.If(poll.OptionThree.Content != "", func() app.UI {
										return app.Div().Body(
											app.Button().Class("deep-orange7 bold white-text responsive").Text(poll.OptionThree.Content).DataSet("option", poll.OptionThree.Content).OnClick(c.onClickPollOption).ID(poll.ID).Name(poll.OptionThree.Content).Disabled(c.pollsButtonDisabled).Style("border-radius", "8px"),
											app.Div().Class("space"),
										)
									}),
								)
								// show results instead
							}).ElseIf(userVoted || poll.Author == c.user.Nickname, func() app.UI {
								return app.Div().Body(

									// voted option I
									app.Div().Class("medium-space border").Style("border-radius", "8px").Body(
										app.Div().Class("bold progress left deep-orange3 medium padding").Style("border-radius", "8px").Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.FormatInt(optionOneShare, 10)+"% 100%, "+strconv.FormatInt(optionOneShare, 10)+"% 0%);"),
										//app.Progress().Value(strconv.Itoa(optionOneShare)).Max(100).Class("deep-orange-text padding medium bold left"),
										//app.Div().Class("progress left light-green"),
										app.Div().Class("middle right-align bold").Body(
											app.Span().Text(poll.OptionOne.Content+" ("+strconv.FormatInt(optionOneShare, 10)+"%)"),
										),
									),

									app.Div().Class("medium-space"),

									// voted option II
									app.Div().Class("medium-space border").Style("border-radius", "8px").Body(
										app.Div().Class("bold progress left deep-orange5 medium padding").Style("border-radius", "8px").Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.FormatInt(optionTwoShare, 10)+"% 100%, "+strconv.FormatInt(optionTwoShare, 10)+"% 0%);").Body(),
										//app.Progress().Value(strconv.Itoa(optionTwoShare)).Max(100).Class("deep-orange-text padding medium bold left"),
										app.Div().Class("middle right-align bold").Body(
											app.Span().Text(poll.OptionTwo.Content+" ("+strconv.FormatInt(optionTwoShare, 10)+"%)"),
										),
									),

									app.Div().Class("space"),

									// voted option III
									app.If(poll.OptionThree.Content != "", func() app.UI {
										return app.Div().Body(
											app.Div().Class("space"),
											app.Div().Class("medium-space border").Style("border-radius", "8px").Body(
												app.Div().Class("bold progress left deep-orange9 medium padding").Style("border-radius", "8px").Style("clip-path", "polygon(0% 0%, 0% 100%, "+strconv.FormatInt(optionThreeShare, 10)+"% 100%, "+strconv.FormatInt(optionThreeShare, 10)+"% 0%);"),
												//app.Progress().Value(strconv.Itoa(optionThreeShare)).Max(100).Class("deep-orange-text deep-orange padding medium bold left"),
												app.Div().Class("middle bold right-align").Body(
													app.Span().Text(poll.OptionThree.Content+" ("+strconv.FormatInt(optionThreeShare, 10)+"%)"),
												),
											),

											app.Div().Class("space"),
										)
									}),
								)
							}),

							// bottom row of the poll
							app.Div().Class("row").Body(
								app.Div().Class("max").Body(
									//app.Text(poll.Timestamp.Format("Jan 02, 2006; 15:04:05")),
									app.Text(pollTimestamp),
								),
								app.If(poll.Author == c.user.Nickname, func() app.UI {
									return app.Div().Body(
										app.B().Title("vote count").Text(len(poll.Voted)),
										app.Button().Title("delete this poll").ID(key).Class("transparent circle").OnClick(c.onClickDeleteButton).Body(
											app.I().Text("delete"),
										),
									)
								}).Else(func() app.UI {
									return app.Div().Body(
										app.B().Title("vote count").Text(len(poll.Voted)),
										app.Button().Title("just voting allowed").ID(key).Class("transparent circle").Disabled(true).Body(
											app.I().Text("how_to_vote"),
										),
									)
								}),
							),
						),
					)
				}),
			),
		),
		app.Div().ID("page-end-anchor"),
		app.If(c.loaderShow, func() app.UI {
			return app.Div().Body(
				app.Div().Class("small-space"),
				app.Progress().Class("circle center large deep-orange-border active"),
			)
		}),
	)
}
