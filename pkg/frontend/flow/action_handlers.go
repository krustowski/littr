package flow

import (
	"log"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
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

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.NewAction("dismiss")

		interactedPost := c.posts[key]

		if interactedPost.Nickname != c.user.Nickname {
			toast.Text(common.ERR_POST_UNAUTH_DELETE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/posts/" + key,
			Data:        nil,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.posts, key)
		})

		ctx.Defer(func(ctx app.Context) {
			toast.Text(common.MSG_DELETE_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch()
		})
	})
}

func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	// change title back to the clean one
	/*title := app.Window().Get("document")
	if !title.IsNull() && strings.Contains(title.Get("title").String(), "(*)") {
		prevTitle := title.Get("title").String()
		title.Set("title", prevTitle[4:])
	}*/

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
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	// Fetch the very image element/node.
	img := app.Window().GetElementByID(id)
	if img.IsNull() {
		return
	}

	src := img.Get("src").String()

	split := strings.Split(src, ".")
	ext := split[len(split)-1]

	name := strings.TrimLeft(id, "img-")

	// image preview (thumbnail) to the actual image logic
	if (ext != "gif" && strings.Contains(src, "thumb")) || (ext == "gif" && strings.Contains(src, "click")) {
		img.Set("src", "/web/pix/"+name+"."+ext)
		//ctx.JSSrc().Set("style", "max-height: 90vh; max-height: 100%; transition: max-height 0.1s; z-index: 1; max-width: 100%; background-position: center")
		img.Set("style", "max-height: 90vh; transition: max-height 0.1s; z-index: 5; max-width: 100%; background-position: center")
	} else if ext == "gif" && !strings.Contains(src, "thumb") {
		img.Set("src", "/web/click-to-see.gif")
		img.Set("style", "z-index: 1; max-height: 100%; max-width: 100%")
	} else {
		img.Set("src", "/web/pix/thumb_"+name+"."+ext)
		img.Set("style", "z-index: 1; max-height: 100%; max-width: 100%")
	}
}

func (c *Content) handleLink(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host
	path := "/flow/posts/"

	if _, err := strconv.ParseFloat(id, 64); err != nil {
		path = "/flow/users/"
	}

	// Write the link to browsers's clipboard.
	navigator := app.Window().Get("navigator")
	if !navigator.IsNull() {
		clipboard := navigator.Get("clipboard")
		if !clipboard.IsNull() && !clipboard.IsUndefined() {
			clipboard.Call("writeText", scheme+"://"+host+path+id)
		}
	}
	ctx.Navigate(path + id)
}

func (c *Content) handleModalPostDeleteShow(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPostKey = id
		c.deleteModalButtonsDisabled = false
		c.deletePostModalShow = true
		c.postButtonsDisabled = false
		c.buttonDisabled = true
	})
}

func (c *Content) handleModalPostReplyShow(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPostKey = id
		c.modalReplyActive = true
		c.postButtonsDisabled = false
		c.buttonDisabled = true
	})

	ctx.Defer(func(app.Context) {
		app.Window().Get("document").Call("getElementById", "reply-textarea").Call("focus")
	})
}

func (c *Content) handleMouseEnter(_ app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	elem := app.Window().GetElementByID(id)
	if elem.IsNull() {
		return
	}

	elem.Get("style").Call("setProperty", "font-size", "1.2rem")
}

func (c *Content) handleMouseLeave(_ app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	elem := app.Window().GetElementByID(id)
	if elem.IsNull() {
		return
	}

	elem.Get("style").Call("setProperty", "font-size", "1rem")
}

func (c *Content) handleRefresh(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.buttonDisabled = true
		c.postButtonsDisabled = true
		c.refreshClicked = true
	})

	ctx.NewAction("clear")

	/*key, ok := a.Value.(string)
	if !ok {
		key = ""
		//return
	}

	if key == "x" || key == "X" {
		c.hideReplies = !c.hideReplies
	}*/

	ctx.Async(func() {
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
			if posts != nil {
				c.posts = *posts
			}
			if users != nil {
				c.users = *users
				c.user = (*users)[c.key]
			}

			c.loaderShow = false
			c.loaderShowImage = false
			c.buttonDisabled = false
			c.refreshClicked = false
			c.postButtonsDisabled = false
			c.contentLoadFinished = true

			c.toast.TText = ""
		})
	})
}

func (c *Content) handleReply(ctx app.Context, a app.Action) {
	// Prevent double-posting.
	if c.postButtonsDisabled {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.modalReplyActive = true
		c.postButtonsDisabled = true
		c.buttonDisabled = true
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.postButtonsDisabled = false
		})

		postType := "post"

		// trim the spaces on the extremites
		replyPost := strings.TrimSpace(c.replyPostContent)

		if !app.Window().GetElementByID("reply-textarea").IsNull() {
			replyPostFull := strings.TrimSpace(app.Window().GetElementByID("reply-textarea").Get("value").String())

			if len(replyPost) < len(replyPostFull) {
				replyPost = replyPostFull
			}
		}

		// allow picture-only posting
		if replyPost == "" && c.newFigFile == "" {
			toast.Text(common.ERR_INVALID_REPLY).Type(common.TTYPE_ERR).Dispatch()
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				c.interactedPostKey = ""
				c.replyPostContent = ""
				c.newFigData = []byte{}
				c.newFigFile = ""
			})
			return
		}

		if output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		payloadNotif := struct {
			OriginalPost string `json:"post_id"`
		}{
			OriginalPost: c.interactedPostKey,
		}

		input = &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/push",
			Data:        payloadNotif,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		output2 := &common.Response{}

		// create a notification
		if ok := common.FetchData(input, output2); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
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

		// Delete the draft(s) from LocalStorage.
		ctx.LocalStorage().Set("newReplyDraft", nil)
		ctx.LocalStorage().Set("newReplyFigFile", nil)
		ctx.LocalStorage().Set("newReplyFigData", nil)

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

		ctx.Defer(func(ctx app.Context) {
			toast.Text(common.MSG_REPLY_ADDED).Type(common.TTYPE_SUCCESS).Dispatch()
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
		var fqz int64 = 1
		now := time.Now().Unix()
		if now-c.lastFire < 1/fqz {
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

/*func (c *Content) onClickStar(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("star", key)
}*/

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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch()
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

func (c *Content) handleTextareaBlur(ctx app.Context, a app.Action) {
	// Save a new post draft, if the focus on textarea is lost.
	ctx.LocalStorage().Set("newReplyDraft", ctx.JSSrc().Get("value").String())
}

func (c *Content) handleUser(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Navigate("/flow/users/" + id)
}
