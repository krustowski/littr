package flow

import (
	"log"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type stub struct{}

func (c *Content) handleClear(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.loaderShow = true
		c.loaderShowImage = true
		c.contentLoadFinished = false
		c.refreshClicked = true
		c.postButtonsDisabled = true
		//c.pageNoToFetch = 0

		c.posts = nil
		c.users = nil
	})
	return
}

func (c *Content) handleDelete(ctx app.Context, a app.Action) {
	// fetch action's value
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	// nasty
	c.postKey = key

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		key := c.postKey
		interactedPost := c.posts[key]

		if interactedPost.Nickname != c.user.Nickname {
			toast.Text("you only can delete your own posts!").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.deletePostModalShow = false
				c.deleteModalButtonsDisabled = false
			})
		}

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/posts/" + interactedPost.ID,
			Data:        interactedPost,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text("cannot reach backend").Type("error").Dispatch(c, dispatch)
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.posts, key)

			c.deletePostModalShow = false
			c.deleteModalButtonsDisabled = false
		})
	})
}

func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	// change title back to the clean one
	title := app.Window().Get("document")
	if !title.IsNull() && strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", prevTitle[4:])
	}

	ctx.Dispatch(func(ctx app.Context) {
		// hotfix which ensures reply modal is not closed if there is also a snackbar/toast active
		//if !toastShow && c.modalReplyActive {
		if c.toast.TText == "" && c.modalReplyActive {
			c.modalReplyActive = false
		}

		c.toast.TText = ""
		c.toast.TType = ""

		c.escapePressed = false
		c.buttonDisabled = false
		c.postButtonsDisabled = false
		c.deletePostModalShow = false
	})
}

func (c *Content) handleImage(ctx app.Context, a app.Action) {
	ctx.JSSrc().Set("src", "")
}

func (c *Content) handleReply(ctx app.Context, a app.Action) {
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		postType := "post"

		// trim the spaces on the extremites
		replyPost := c.replyPostContent

		if !app.Window().GetElementByID("reply-textarea").IsNull() {
			replyPostFull := strings.TrimSpace(app.Window().GetElementByID("reply-textarea").Get("value").String())

			if len(replyPost) < len(replyPostFull) {
				replyPost = replyPostFull
			}
		}

		// allow picture-only posting
		if replyPost == "" && c.newFigFile == "" {
			toast.Text("no valid reply entered").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = false
			})
			return
		}

		//newPostID := time.Now()
		//stringID := strconv.FormatInt(newPostID.UnixNano(), 10)

		path := "/api/v1/posts"

		// TODO: the Post data model has to be changed
		// migrate Post.ReplyID (int) to Post.ReplyID (string)
		// ReplyID is to be string key to easily refer to other post
		payload := models.Post{
			//ID:        stringID,
			Nickname:  c.user.Nickname,
			Type:      postType,
			Content:   replyPost,
			ReplyToID: c.interactedPostKey,
			Data:      c.newFigData,
			Figure:    c.newFigFile,
			//Timestamp: newPostID,
			//ReplyTo: replyID, <--- is type int
		}

		input := &common.CallInput{
			Method:      "POST",
			Url:         path,
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		type dataModel struct {
			Posts map[string]models.Post `posts`
		}

		output := &common.Response{Data: &dataModel{}}

		if ok := common.FetchData(input, output); !ok {
			toast.Text("cannot reach backend").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = false

				c.interactedPostKey = ""
				c.replyPostContent = ""
				c.newFigData = []byte{}
				c.newFigFile = ""
			})
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text("cannot get data").Type("error").Dispatch(c, dispatch)
			return
		}

		payloadNotif := struct {
			OriginalPost string `json:"original_post"`
		}{
			OriginalPost: c.interactedPostKey,
		}

		input = &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/push/notification/" + c.interactedPostKey,
			Data:        payloadNotif,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		output2 := &common.Response{}

		// create a notification
		if ok := common.FetchData(input, output2); !ok {
			toast.Text("cannot reach backend").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = false

				c.interactedPostKey = ""
				c.replyPostContent = ""
				c.newFigData = []byte{}
				c.newFigFile = ""
			})
			return
		}

		posts := c.posts

		// we do not know the ID, as it is assigned in the BE logic,
		// so we need to loop over the list of posts (1)...
		for k, p := range data.Posts {
			posts[k] = p
		}

		ctx.Dispatch(func(ctx app.Context) {
			// add new post to post list on frontend side to render
			//c.posts[stringID] = payload
			c.posts = posts

			c.modalReplyActive = false
			c.postButtonsDisabled = false
			c.buttonDisabled = false

			c.interactedPostKey = ""
			c.replyPostContent = ""
			c.newFigData = []byte{}
			c.newFigFile = ""
		})
	})
}

func (c *Content) handleScroll(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		// limit the fire rate to 1 Hz
		now := time.Now().Unix()
		if now-c.lastFire < 1/1 {
			return
		}

		if bottom-height < 0 && !c.paginationEnd && !c.processingFire && !c.loaderShow {
			ctx.Dispatch(func(ctx app.Context) {
				c.loaderShow = true
				c.processingFire = true
				c.contentLoadFinished = false
			})

			var newPosts *map[string]models.Post
			var newUsers *map[string]models.User

			posts := c.posts
			users := c.users

			updated := false
			lastPageFetched := c.lastPageFetched

			// fetch more posts
			//if (c.pageNoToFetch+1)*(c.pagination*2) >= len(posts) && !lastPageFetched {
			if !lastPageFetched {
				opts := pageOptions{
					PageNo:   c.pageNoToFetch,
					Context:  ctx,
					CallerID: c.user.Nickname,

					//SinglePost:   parts.SinglePost,
					SinglePost: c.singlePostID != "",
					//SinglePostID: parts.SinglePostID,
					SinglePostID: c.singlePostID,
					//UserFlow:     parts.UserFlow,
					UserFlow: c.userFlowNick != "",
					//UserFlowNick: parts.UserFlowNick,
					UserFlowNick: c.userFlowNick,

					Hashtag:     c.hashtag,
					HideReplies: c.hideReplies,
				}

				newPosts, newUsers = c.fetchFlowPage(opts)
				postControlCount := len(posts)

				// patch single-post and user flow atypical scenarios
				if posts == nil {
					posts = make(map[string]models.Post)
				}
				if users == nil {
					users = make(map[string]models.User)
				}

				// append/insert more posts/users
				for key, post := range *newPosts {
					posts[key] = post
				}
				for key, user := range *newUsers {
					users[key] = user
				}

				updated = true

				// no more posts, fetching another page does not make sense
				if len(posts) == postControlCount {
					//updated = false
					lastPageFetched = true

				}
			}

			ctx.Dispatch(func(ctx app.Context) {
				c.lastFire = now
				c.pageNoToFetch++
				c.pageNo++

				if updated {
					c.posts = posts
					c.users = users

					log.Println("updated")
				}

				c.processingFire = false
				c.loaderShow = false
				c.contentLoadFinished = true
				c.lastPageFetched = lastPageFetched

				//log.Println("new content page request fired")
			})
			return
		}
	})
}

func (c *Content) handleRefresh(ctx app.Context, a app.Action) {
	// little hack to dismiss navbar's snackbar
	snack := app.Window().GetElementByID("snackbar-general")
	if !snack.IsNull() {
		snack.Get("classList").Call("remove", "active")
	}

	ctx.Async(func() {
		// nasty hotfix, TODO
		//c.pageNoToFetch = 0

		//parts := c.parseFlowURI(ctx)

		opts := pageOptions{
			//PageNo:   c.pageNoToFetch,
			PageNo:   0,
			Context:  ctx,
			CallerID: c.user.Nickname,

			//SinglePost:   parts.SinglePost,
			SinglePost: c.singlePostID != "",
			//SinglePostID: parts.SinglePostID,
			SinglePostID: c.singlePostID,
			//UserFlow:     parts.UserFlow,
			UserFlow: c.userFlowNick != "",
			//UserFlowNick: parts.UserFlowNick,
			UserFlowNick: c.userFlowNick,
			Hashtag:      c.hashtag,
			HideReplies:  c.hideReplies,
		}

		posts, users := c.fetchFlowPage(opts)

		ctx.Dispatch(func(ctx app.Context) {
			c.posts = *posts
			c.users = *users

			c.user = (*users)[c.key]

			c.loaderShow = false
			c.loaderShowImage = false
			c.refreshClicked = false
			c.postButtonsDisabled = false
			c.contentLoadFinished = true

			c.toast.TText = ""
		})
	})
}
func (c *Content) handleStar(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	// runs on the main UI goroutine via a component ActionHandler
	post := c.posts[key]
	post.ReactionCount++
	c.posts[key] = post
	c.postKey = key

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		//key := ctx.JSSrc().Get("id").String()
		key := c.postKey
		//author = c.user.Nickname

		interactedPost := c.posts[key]
		//interactedPost.ReactionCount++

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/posts/" + interactedPost.ID + "/star",
			Data:        interactedPost,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		type dataModel struct {
			Posts map[string]models.Post `json:"posts"`
		}

		output := &common.Response{Data: &dataModel{}}

		// add new post to backend struct
		if ok := common.FetchData(input, output); !ok {
			toast.Text("cannot reach backend").Type("error").Dispatch(c, dispatch)
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text("cannot get data").Type("error").Dispatch(c, dispatch)
			return
		}

		// keep the reply count!
		oldPost := c.posts[key]
		newPost := data.Posts[key]
		newPost.ReplyCount = oldPost.ReplyCount

		ctx.Dispatch(func(ctx app.Context) {
			c.posts[key] = newPost
		})
	})
}
