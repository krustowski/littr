package polls

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// handleDelete()
func (c *Content) handleDelete(ctx app.Context, a app.Action) {
	// get and cast the action value
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	// a fuse to ensure that clicked poll's button is the one
	// the event and action bears
	if key != c.interactedPollKey {
		return
	}

	// fetch the struct of page's toast
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		//key := c.pollKey
		interactedPoll := c.polls[key]

		if interactedPoll.Author != c.user.Nickname {
			toast.Text(common.ERR_POLL_UNAUTH_DELETE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.deletePollModalShow = false
				c.deleteModalButtonsDisabled = false
			})
			return
		}

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/polls/" + interactedPoll.ID,
			Data:        interactedPoll,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.polls, key)

			c.deletePollModalShow = false
			c.deleteModalButtonsDisabled = false
		})
	})
}

// handleDismiss()
func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toast.TText = ""

		c.pollsButtonDisabled = false
		c.deletePollModalShow = false
	})
}

// handleScroll()
func (c *Content) handleScroll(ctx app.Context, a app.Action) {
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		if bottom-height < 0 && !c.paginationEnd && !c.processingScroll {
			ctx.Dispatch(func(ctx app.Context) {
				c.processingScroll = true
			})

			pageNo := c.pageNo

			input := &common.CallInput{
				Method: "GET",
				Url:    "/api/v1/polls",
				Data:   nil,
				PageNo: pageNo,
			}

			type dataModel struct {
				Polls map[string]models.Poll `json:"polls"`
				User  models.User            `json:"user"`
			}

			output := &common.Response{Data: &dataModel{}}

			// call the API to fetch the data
			if ok := common.FetchData(input, output); !ok {
				toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			if output.Code == 401 {
				toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Dispatch(c, dispatch)
				return
			}

			if output.Code != 200 {
				toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			data, ok := output.Data.(*dataModel)
			if !ok {
				toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			polls := c.polls
			if polls == nil {
				polls = make(map[string]models.Poll)
			}

			for key, poll := range data.Polls {
				polls[key] = poll
			}

			ctx.Dispatch(func(ctx app.Context) {
				c.pageNo++
				c.polls = polls
				c.processingScroll = false
			})
			return
		}
	})
}

// handleVote()
func (c *Content) handleVote(ctx app.Context, a app.Action) {
	// fetch the action's value
	keys, ok := a.Value.([]string)
	if !ok {
		return
	}

	key := keys[0]
	option := keys[1]

	poll := c.polls[key]
	toast := common.Toast{AppContext: &ctx}

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
		toast.Text(common.ERR_POLL_OPTION_MISMATCH).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
	}

	ctx.Async(func() {
		input := &common.CallInput{
			Method:      "PUT",
			Url:         "/api/v1/polls/" + poll.ID,
			Data:        poll,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.polls[key] = poll
			c.pollsButtonDisabled = false
		})
	})
}
