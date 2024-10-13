package post

import (
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) onClick(ctx app.Context, e app.Event) {
	// prevent double-posting
	if c.postButtonsDisabled {
		return
	}

	// post, fig, poll
	postType := ctx.JSSrc().Get("id").String()
	content := ""
	poll := models.Poll{}

	toast := common.Toast{AppContext: &ctx}

	var payload interface{}

	c.postButtonsDisabled = true

	ctx.Async(func() {
		switch postType {
		case "poll":
			// trim the padding spaces on the extremities
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			pollQuestion := strings.TrimSpace(c.pollQuestion)
			pollOptionI := strings.TrimSpace(c.pollOptionI)
			pollOptionII := strings.TrimSpace(c.pollOptionII)
			pollOptionIII := strings.TrimSpace(c.pollOptionIII)

			if pollOptionI == "" || pollOptionII == "" || pollQuestion == "" {
				toast.Text(common.ERR_POLL_FIELDS_REQUIRED).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				break
			}

			now := time.Now()
			content = strconv.FormatInt(now.UnixNano(), 10)

			poll.ID = content
			poll.Question = pollQuestion
			poll.OptionOne.Content = pollOptionI
			poll.OptionTwo.Content = pollOptionII
			poll.OptionThree.Content = pollOptionIII
			poll.Timestamp = now
			break

		case "post":
			textarea := app.Window().GetElementByID("post-textarea").Get("value").String()

			// trim the padding spaces on the extremities
			// https://www.tutorialspoint.com/how-to-trim-a-string-in-golang
			//newPost := strings.TrimSpace(c.newPost)
			newPost := strings.TrimSpace(textarea)

			// allow just picture posting
			if newPost == "" && c.newFigFile == "" {
				toast.Text(common.ERR_POST_TEXTAREA_EMPTY).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				break
			}

			content = newPost
			break

		default:
			break
		}

		// fixing the bug: cannot use return and break together in switch according to Sonar
		if c.toast.TText != "" {
			return
		}

		var encoded string
		var user models.User

		ctx.LocalStorage().Get("user", &encoded)

		// decode and unmarshal the local storage user data
		if err := common.LoadUser(encoded, &user); err != nil {
			toast.Text(common.ERR_LOCAL_STORAGE_LOAD_FAIL).Error(err).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		path := "/api/v1/posts"

		if postType == "post" {
			payload = models.Post{
				Nickname: user.Nickname,
				Type:     postType,
				Content:  content,
				PollID:   poll.ID,
				Figure:   c.newFigFile,
				Data:     c.newFigData,
				//Timestamp: time.Now(),
			}
		} else if postType == "poll" {
			path = "/api/v1/polls"
			poll.Author = user.Nickname
			payload = poll
		}

		input := &common.CallInput{
			Method:      "POST",
			Url:         path,
			Data:        payload,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		// add new post/poll to backend struct
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE.Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if postType == "poll" {
			ctx.Navigate("/polls")
		} else {
			ctx.Navigate("/flow")
		}
	})
}

func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	textarea := app.Window().GetElementByID("post-textarea")

	//if textarea.Get("value").IsNull() {
	if textarea.IsNull() {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" && len(textarea.Get("value").String()) != 0 {
		app.Window().GetElementByID("post").Call("click")
	}
}

// https://github.com/maxence-charriere/go-app/issues/882
func (c *Content) handleFigUpload(ctx app.Context, e app.Event) {
	file := e.Get("target").Get("files").Index(0)

	//log.Println("name", file.Get("name").String())
	//log.Println("size", file.Get("size").Int())
	//log.Println("type", file.Get("type").String())

	c.postButtonsDisabled = true

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		if data, err := common.ReadFile(file); err != nil {
			toast.Error(err).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return

		} else {
			/*payload := models.Post{
				Nickname:  author,
				Type:      "fig",
				Content:   file.Get("name").String(),
				Timestamp: time.Now(),
				Data:      data,
			}*/

			// add new post/poll to backend struct
			/*if _, ok := littrAPI("POST", path, payload, user.Nickname, 0); !ok {
				toastText = "backend error: cannot add new content"
				log.Println("cannot post new content to API!")
			} else {
				ctx.Navigate("/flow")
			}*/

			toast.Text(common.MSG_IMAGE_READY).Type(common.TTYPE_INFO).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.newFigFile = file.Get("name").String()
				c.newFigData = data
			})
			return

		}
	})
}
