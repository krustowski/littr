package post

import (
	"strconv"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"
)

func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toast.TText = ""
		c.toast.TType = ""
		c.postButtonsDisabled = false
	})
}

// onClick is a callback function to post generic input (post, or poll).
func (c *Content) handlePostPoll(ctx app.Context, a app.Action) {
	bt, ok := a.Value.(string)
	if !ok {
		return
	}

	var postType string

	// Fetch the type: post, fig, or poll.
	switch bt {
	case "button-new-post":
		postType = "post"

	case "button-new-poll":
		postType = "poll"
	}

	// Prevent the double-posting.
	if c.postButtonsDisabled {
		return
	}

	// Null the content (field of the payload, read on).
	content := ""
	poll := models.Poll{}

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	var payload interface{}

	// Nasty way on how to disable the buttons. (Use action handler and Dispatch function instead.)
	c.postButtonsDisabled = true

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.postButtonsDisabled = false
		})

		var leave bool

		// Determine the post Type.
		switch postType {
		case "poll":
			// Trim the padding spaces on the extremities.
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			pollQuestion := strings.TrimSpace(c.pollQuestion)
			pollOptionI := strings.TrimSpace(c.pollOptionI)
			pollOptionII := strings.TrimSpace(c.pollOptionII)
			pollOptionIII := strings.TrimSpace(c.pollOptionIII)

			if pollOptionI == "" || pollOptionII == "" || pollQuestion == "" {
				toast.Text(common.ERR_POLL_FIELDS_REQUIRED).Type(common.TTYPE_ERR).Dispatch()
				leave = true
				break
			}

			if pollOptionIII == "" {
				pollOptionIII = strings.TrimSpace(app.Window().GetElementByID("poll-option-iii").Get("value").String())
			}

			// Compose a timestamp and the derived key (content).
			now := time.Now()
			content = strconv.FormatInt(now.UnixNano(), 10)

			// Assign various poll's field inputs to the generic poll.
			poll.ID = content
			poll.Question = pollQuestion
			poll.OptionOne.Content = pollOptionI
			poll.OptionTwo.Content = pollOptionII
			poll.OptionThree.Content = pollOptionIII
			poll.Timestamp = now

		case "post":
			// This is to hotfix the fact that the input can be CTRL-Entered in.
			textarea := app.Window().GetElementByID("post-textarea").Get("value").String()

			// Trim the padding spaces on the extremities.
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			newPost := strings.TrimSpace(textarea)

			// Allow a just picture posting.
			if newPost == "" && c.newFigFile == "" {
				toast.Text(common.ERR_POST_TEXTAREA_EMPTY).Type(common.TTYPE_ERR).Dispatch()
				leave = true
				break
			}

			// Content is the new post itself.
			content = newPost

		default:
			toast.Text(common.ERR_POST_UNKNOWN_TYPE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Fixing the bug: cannot use return and break together in switch according to Sonar.
		if leave {
			return
		}

		var user models.User

		// Decode and unmarshal the local storage user data.
		if err := common.LoadUser(&user, &ctx); err != nil {
			toast.Text(common.ERR_LOCAL_STORAGE_LOAD_FAIL).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		path := "/api/v1/posts"

		// Compose a post payload.
		switch postType {
		case "post":
			payload = models.Post{
				Nickname: user.Nickname,
				Type:     postType,
				Content:  content,
				PollID:   poll.ID,
				Figure:   c.newFigFile,
				Data:     c.newFigData,
			}
		case "poll":
			// Compose a poll payload.
			path = "/api/v1/polls"
			poll.Author = user.Nickname

			payload = struct {
				Question    string `json:"question"`
				OptionOne   string `json:"option_one"`
				OptionTwo   string `json:"option_two"`
				OptionThree string `json:"option_three"`
			}{
				Question:    poll.Question,
				OptionOne:   poll.OptionOne.Content,
				OptionTwo:   poll.OptionTwo.Content,
				OptionThree: poll.OptionThree.Content,
			}
		}

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "POST",
			Url:         path,
			Data:        payload,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output := &common.Response{}

		// Send new post/poll to backend struct.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Check for the HTTP 201 response code, otherwise print the API response message in the toast.
		if output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Delete the draft(s) from LocalStorage.
		ctx.LocalStorage().Set("newPostDraft", nil)
		ctx.LocalStorage().Set("newPostFigFile", nil)
		ctx.LocalStorage().Set("newPostFigData", nil)

		// Redirection according to the post type.
		if postType == "poll" {
			ctx.Navigate("/polls")
		} else {
			ctx.Navigate("/flow")
		}
	})
}
