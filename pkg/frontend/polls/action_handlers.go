package polls

import (
	"log"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// handleDelete()
func (c *Content) handleDelete(ctx app.Context, a app.Action) {
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

		input := common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/polls/" + interactedPoll.ID,
			Data:        interactedPoll,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: false,
		}

		if _, ok := common.CallAPI(input); !ok {
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

// handleDismiss()
func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
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
				log.Println("new content page request fired")
			})
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

		if _, ok := common.CallAPI(input); !ok {
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
