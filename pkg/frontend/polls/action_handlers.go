package polls

import (
	"log"

	"go.vxn.dev/littr/pkg/frontend/common"

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
	toast := Toast{AppContext: &ctx}

	ctx.Async(func() {
		//key := c.pollKey
		interactedPoll := c.polls[key]

		if interactedPoll.Author != c.user.Nickname {
			toast.Text("you only can delete your own polls!").Dispatch(c)

			ctx.Dispatch(func(ctx app.Context) {
				c.deletePollModalShow = false
				c.deleteModalButtonsDisabled = false
			})
			return
		}

		input := common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/polls/" + interactedPoll.ID,
			Data:        interactedPoll,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: false,
		}

		output := &struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{}

		if ok := common.CallAPI(input, output); !ok {
			toast.Text("backend error: cannot delete a poll").Type("error").Dispatch(c)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c)
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
		c.toast.ToastText = ""
		c.toastText = ""
		c.toastShow = false

		c.pollsButtonDisabled = false
		c.deletePollModalShow = false
	})
}

// handleScroll()
func (c *Content) handleScroll(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		if bottom-height < 0 && !c.paginationEnd {
			ctx.Dispatch(func(ctx app.Context) {
				c.pageNo++
			})

			log.Println("new content page request fired")
			return
		}
	})
}

// handleVote()
func (c *Content) handleVote(ctx app.Context, a app.Action) {
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

		input := common.CallInput{
			Method:      "PUT",
			Url:         "/api/v1/polls/" + poll.ID,
			Data:        poll,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		if ok := common.CallAPI(input, &struct{}{}); !ok {
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
