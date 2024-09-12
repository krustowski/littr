package frontend

import (
	"encoding/json"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type FlowPage struct {
	app.Compo
}

type flowContent struct {
	app.Compo

	loaderShow      bool
	loaderShowImage bool

	hideReplies bool

	contentLoadFinished bool

	loggedUser string
	user       models.User
	key        string

	toastShow bool
	toastText string
	toastType string

	buttonDisabled      bool
	postButtonsDisabled bool
	modalReplyActive    bool
	replyPostContent    string
	newFigLink          string
	newFigFile          string
	newFigData          []byte

	escapePressed bool

	deletePostModalShow        bool
	deleteModalButtonsDisabled bool

	interactedPostKey string
	singlePostID      string
	isPost            bool
	userFlowNick      string

	paginationEnd  bool
	pagination     int
	pageNo         int
	pageNoToFetch  int
	lastFire       int64
	processingFire bool

	lastPageFetched bool

	postKey     string
	posts       map[string]models.Post
	users       map[string]models.User
	sortedPosts []models.Post

	refreshClicked bool

	hashtag string

	eventListener        func()
	eventListenerMsg     func()
	keyDownEventListener func()
	dismissEventListener func()
}

func (p *FlowPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("flow / littr")
}

func (p *FlowPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&flowContent{},
	)
}

func (c *flowContent) onClickFollow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	if key == c.user.Nickname {
		return
	}

	flowList := c.user.FlowList

	if c.user.ShadeList[key] {
		return
	}

	if flowList == nil {
		flowList = make(map[string]bool)
		flowList[c.user.Nickname] = true
		//c.user.FlowList = flowList
	}

	if value, found := flowList[key]; found {
		if !value && c.users[key].Private {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonDisabled = false
				c.postButtonsDisabled = false
				c.toastText = "this account is private"
				c.toastShow = true
			})

			return
		}
		flowList[key] = !flowList[key]
	} else {
		if c.users[key].Private {
			ctx.Dispatch(func(ctx app.Context) {
				c.buttonDisabled = false
				c.postButtonsDisabled = false
				c.toastText = "this account is private"
				c.toastShow = true
			})
			return
		}
		flowList[key] = true
	}

	ctx.Async(func() {
		ctx.Dispatch(func(ctx app.Context) {
			c.buttonDisabled = true
			c.postButtonsDisabled = true
		})

		toastText := ""

		payload := struct {
			FlowList map[string]bool `json:"flow_list"`
		}{
			FlowList: flowList,
		}

		input := callInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: c.hideReplies,
		}

		respRaw, ok := littrAPI(input)
		if !ok {
			toastText = "generic backend error"
			return
		}

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if err := json.Unmarshal(*respRaw, &response); err != nil {
			toastText = "user update failed: " + err.Error()
			return
		}

		if response.Code != 200 && response.Code != 201 {
			toastText = "user update failed: " + response.Message
			log.Println(response.Message)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")

			c.buttonDisabled = false
			c.postButtonsDisabled = false

			c.user.FlowList = flowList
		})
	})
}

func (c *flowContent) onClickLink(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host

	// write the link to browsers's clipboard
	navigator := app.Window().Get("navigator")
	if !navigator.IsNull() {
		clipboard := navigator.Get("clipboard")
		if !clipboard.IsNull() && !clipboard.IsUndefined() {
			clipboard.Call("writeText", scheme+"://"+host+"/flow/post/"+key)
		}
	}
	ctx.Navigate("/flow/post/" + key)
}

func (c *flowContent) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		// hotfix which ensures reply modal is not closed if there is also a snackbar/toast active
		//if !toastShow && c.modalReplyActive {
		if !c.toastShow && c.modalReplyActive {
			c.modalReplyActive = false
		}

		c.escapePressed = false
		c.toastShow = false
		c.toastText = ""
		c.toastType = ""
		c.buttonDisabled = false
		c.postButtonsDisabled = false
		c.deletePostModalShow = false
	})
}

func (c *flowContent) onClickDismiss(ctx app.Context, e app.Event) {
	//ctx.NewActionWithValue("dismiss", key)
	ctx.NewAction("dismiss")

	//ctx.Dispatch(func(ctx app.Context) {
	/*if c.escapePressed {
		// hotfix which ensures reply modal is not closed if there is also a snackbar/toast active
		if c.toastText == "" || c.escapePressed {
			c.modalReplyActive = false
		}

		c.escapePressed = false
		c.toastShow = false
		c.toastText = ""
		c.toastType = ""
		c.buttonDisabled = false
		c.postButtonsDisabled = false
		c.deletePostModalShow = false
		//})
	}*/
}

// https://github.com/maxence-charriere/go-app/issues/882
func (c *flowContent) handleFigUpload(ctx app.Context, e app.Event) {
	var toastText string

	file := e.Get("target").Get("files").Index(0)

	//log.Println("name", file.Get("name").String())
	//log.Println("size", file.Get("size").Int())
	//log.Println("type", file.Get("type").String())

	c.postButtonsDisabled = true

	ctx.Async(func() {
		if data, err := readFile(file); err != nil {
			toastText = err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = true
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.postButtonsDisabled = false
			})
			return

		} else {
			toastText = "image is ready"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastType = "success"
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.postButtonsDisabled = false

				c.newFigFile = file.Get("name").String()
				c.newFigData = data
			})
			return

		}
	})
}

func (c *flowContent) onClickImage(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	src := ctx.JSSrc().Get("src").String()

	split := strings.Split(src, ".")
	ext := split[len(split)-1]

	// image preview (thumbnail) to the actual image logic
	if (ext != "gif" && strings.Contains(src, "thumb")) || (ext == "gif" && strings.Contains(src, "click")) {
		ctx.JSSrc().Set("src", "/web/pix/"+key+"."+ext)
		//ctx.JSSrc().Set("style", "max-height: 90vh; max-height: 100%; transition: max-height 0.1s; z-index: 1; max-width: 100%; background-position: center")
		ctx.JSSrc().Set("style", "max-height: 90vh; transition: max-height 0.1s; z-index: 5; max-width: 100%; background-position")
	} else if ext == "gif" && !strings.Contains(src, "thumb") {
		ctx.JSSrc().Set("src", "/web/click-to-see.gif")
		ctx.JSSrc().Set("style", "z-index: 1; max-height: 100%; max-width: 100%")
	} else {
		ctx.JSSrc().Set("src", "/web/pix/thumb_"+key+"."+ext)
		ctx.JSSrc().Set("style", "z-index: 1; max-height: 100%; max-width: 100%")
	}
}

func (c *flowContent) handleImage(ctx app.Context, a app.Action) {
	ctx.JSSrc().Set("src", "")
}

func (c *flowContent) onClickUserFlow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	//c.buttonDisabled = true

	ctx.Navigate("/flow/user/" + key)
}

// onClickReply acts like a caller function evoked when user click on the reply icon at one's post
func (c *flowContent) onClickReply(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPostKey = ctx.JSSrc().Get("id").String()
		c.modalReplyActive = true
		c.postButtonsDisabled = false
		c.buttonDisabled = true
	})
}

// onClickPostReply acts like a caller function evoked when user clicks on "reply" button in the reply modal
func (c *flowContent) onClickPostReply(ctx app.Context, e app.Event) {
	// prevent double-posting
	if c.postButtonsDisabled {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.modalReplyActive = true
		c.postButtonsDisabled = true
		c.buttonDisabled = true
	})

	ctx.NewAction("reply")
}

func (c *flowContent) handleReply(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		toastText := ""

		// TODO: allow figs in replies
		// check if the contents is a valid URL, then change the type to "fig"
		postType := "post"

		// trim the spaces on the extremites
		replyPost := c.replyPostContent

		if replyPost == "" && !app.Window().GetElementByID("reply-textarea").IsNull() {
			replyPost = strings.TrimSpace(app.Window().GetElementByID("reply-textarea").Get("value").String())
		}

		// allow picture-only posting
		if replyPost == "" && c.newFigFile == "" {
			toastText = "no valid reply entered"

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = false
				c.toastText = toastText
				c.toastShow = (toastText != "")
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

		postsRaw := struct {
			Posts map[string]models.Post `posts`
		}{}

		input := callInput{
			Method:      "POST",
			Url:         path,
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		// add new post/poll to backend struct
		if resp, _ := littrAPI(input); resp != nil {
			err := json.Unmarshal(*resp, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parsing error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
					c.postButtonsDisabled = false

					c.interactedPostKey = ""
					c.replyPostContent = ""
					c.newFigData = []byte{}
					c.newFigFile = ""
				})
				return
			}
		} else {
			log.Println("cannot fetch post flow list")
			toastText = "API error: cannot fetch the post list"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.postButtonsDisabled = false

				c.interactedPostKey = ""
				c.replyPostContent = ""
				c.newFigData = []byte{}
				c.newFigFile = ""
			})
			return
		}

		payloadNotif := struct {
			OriginalPost string `json:"original_post"`
		}{
			OriginalPost: c.interactedPostKey,
		}

		input = callInput{
			Method:      "POST",
			Url:         "/api/v1/push/notification/" + c.interactedPostKey,
			Data:        payloadNotif,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		// create a notification
		if _, ok := littrAPI(input); !ok {
			toastText = "cannot POST new notification"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
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
		for k, p := range postsRaw.Posts {
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

func (c *flowContent) onClickGeneric(ctx app.Context, e app.Event) {
	if e.Get("target").String() == "overlay" || e.Get("srcElement").String() == "overlay" {
		log.Println("overlay clicked")
	}
}

func (c *flowContent) onKeyDown(ctx app.Context, e app.Event) {
	if (e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc") && !c.escapePressed {
		ctx.NewAction("dismiss")
		return
	}

	textarea := app.Window().GetElementByID("reply-textarea")

	if (e.Get("key").String() == "x" || e.Get("key").String() == "X" || e.Get("key").String() == "r" || e.Get("key").String() == "R") && textarea.IsNull() && !c.refreshClicked {
		ctx.Dispatch(func(ctx app.Context) {
			// experimental feature to toggle show/hide reply posts on flow
			if e.Get("key").String() == "x" || e.Get("key").String() == "X" {
				c.hideReplies = !c.hideReplies
			}

			c.loaderShow = true
			c.loaderShowImage = true
			c.contentLoadFinished = false
			c.refreshClicked = true
			c.postButtonsDisabled = true
			//c.pageNoToFetch = 0

			c.posts = nil
			c.users = nil
		})

		ctx.NewAction("refresh")
		return
	}

	/*
	 *  autosubmit via ctrl+enter
	 */

	if textarea.IsNull() {
		return
	}

	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" && len(textarea.Get("value").String()) != 0 {
		app.Window().GetElementByID("reply").Call("click")
	}
}

func (c *flowContent) onScroll(ctx app.Context, e app.Event) {
	ctx.NewAction("scroll")
}

func (c *flowContent) handleScroll(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		// limit the fire rate to 1 Hz
		now := time.Now().Unix()
		if now-c.lastFire < 1 {
			return
		}

		if bottom-height < 0 && !c.paginationEnd && !c.processingFire && !c.loaderShow {
			ctx.Dispatch(func(ctx app.Context) {
				c.loaderShow = true
				c.processingFire = true
				c.contentLoadFinished = false
			})

			var newPosts map[string]models.Post
			var newUsers map[string]models.User

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
				for key, post := range newPosts {
					posts[key] = post
				}
				for key, user := range newUsers {
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

func (c *flowContent) onClickDeleteButton(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPostKey = key
		c.deleteModalButtonsDisabled = false
		c.deletePostModalShow = true
	})
}

func (c *flowContent) onClickDelete(ctx app.Context, e app.Event) {
	key := c.interactedPostKey
	ctx.NewActionWithValue("delete", key)

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteModalButtonsDisabled = true
	})
}

func (c *flowContent) handleDelete(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	c.postKey = key

	ctx.Async(func() {
		var toastText string = ""

		key := c.postKey
		interactedPost := c.posts[key]

		if interactedPost.Nickname != c.user.Nickname {
			toastText = "you only can delete your own posts!"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.deletePostModalShow = false
				c.deleteModalButtonsDisabled = false
			})
		}

		input := callInput{
			Method:      "DELETE",
			Url:         "/api/v1/posts/" + interactedPost.ID,
			Data:        interactedPost,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		if _, ok := littrAPI(input); !ok {
			toastText = "backend error: cannot delete a post"
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.posts, key)

			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.deletePostModalShow = false
			c.deleteModalButtonsDisabled = false
		})
	})
}

func (c *flowContent) onClickStar(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("star", key)
}

func (c *flowContent) handleStar(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	// runs on the main UI goroutine via a component ActionHandler
	post := c.posts[key]
	post.ReactionCount++
	c.posts[key] = post
	c.postKey = key

	ctx.Async(func() {
		//var author string
		var toastText string = ""

		//key := ctx.JSSrc().Get("id").String()
		key := c.postKey
		//author = c.user.Nickname

		interactedPost := c.posts[key]
		//interactedPost.ReactionCount++

		postsRaw := struct {
			Posts map[string]models.Post `json:"posts"`
		}{}

		input := callInput{
			Method:      "PATCH",
			Url:         "/api/v1/posts/" + interactedPost.ID + "/star",
			Data:        interactedPost,
			CallerID:    c.user.Nickname,
			PageNo:      c.pageNo,
			HideReplies: c.hideReplies,
		}

		// add new post to backend struct
		if resp, ok := littrAPI(input); ok {
			err := json.Unmarshal(*resp, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parsing error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			toastText = "backend error: cannot rate a post"
		}

		// keep the reply count!
		oldPost := c.posts[key]
		newPost := postsRaw.Posts[key]
		newPost.ReplyCount = oldPost.ReplyCount

		ctx.Dispatch(func(ctx app.Context) {
			c.posts[key] = newPost
			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
	})
}

func (c *flowContent) onMessage(ctx app.Context, e app.Event) {
	data := e.JSValue().Get("data").String()
	log.Println("msg event: data:" + data)

	if data == "heartbeat" || data == c.user.Nickname {
		return
	}

	if _, flowed := c.user.FlowList[data]; !flowed {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.toastText = "new post added above"
		c.toastType = "info"
	})
}

func (c *flowContent) handleRefresh(ctx app.Context, a app.Action) {
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
			c.posts = posts
			c.users = users

			c.user = users[c.key]

			c.loaderShow = false
			c.loaderShowImage = false
			c.refreshClicked = false
			c.postButtonsDisabled = false
			c.contentLoadFinished = true

			c.toastText = ""
			c.toastShow = false
		})
	})
}

func (c *flowContent) onClickRefresh(ctx app.Context, e app.Event) {
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

	ctx.NewAction("dismiss")
	ctx.NewAction("refresh")
}

type pageOptions struct {
	PageNo   int `default:0`
	Context  app.Context
	CallerID string

	SinglePost bool `default:"false"`
	UserFlow   bool `default:"false"`

	SinglePostID string `default:""`
	UserFlowNick string `default:""`

	Hashtag string `default:""`

	HideReplies bool `default:"false"`
}

func (c *flowContent) fetchFlowPage(opts pageOptions) (map[string]models.Post, map[string]models.User) {
	var toastText string
	var toastType string

	resp := struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Code  int                    `json:"code"`
		Key   string                 `json:"key"`
	}{}

	ctx := opts.Context
	pageNo := opts.PageNo

	if opts.Context == nil {
		toastText = "app context pointer cannot be nil"
		log.Println(toastText)

		return nil, nil
	}

	//pageNo := c.pageNoToFetch
	if c.refreshClicked {
		pageNo = 0
	}
	//pageNoString := strconv.FormatInt(int64(pageNo), 10)

	url := "/api/v1/posts"
	if opts.UserFlow || opts.SinglePost || opts.Hashtag != "" {
		if opts.SinglePostID != "" {
			url += "/" + opts.SinglePostID
		}

		if opts.UserFlowNick != "" {
			//url += "/user/" + opts.UserFlowNick
			url = "/api/v1/users/" + opts.UserFlowNick + "/posts"
		}

		if opts.Hashtag != "" {
			url = "/api/v1/posts/hashtag/" + opts.Hashtag
		}

		if opts.SinglePostID == "" && opts.UserFlowNick == "" && opts.Hashtag == "" {
			toastText = "page parameters (singlePost, userFlow, hashtag) cannot be blank"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.refreshClicked = false
			})
			return nil, nil
		}

	}

	input := callInput{
		Method:      "GET",
		Url:         url,
		Data:        nil,
		CallerID:    c.user.Nickname,
		PageNo:      pageNo,
		HideReplies: c.hideReplies,
	}

	if byteData, _ := littrAPI(input); byteData != nil {
		err := json.Unmarshal(*byteData, &resp)
		if err != nil {
			log.Println(err.Error())
			toastText = "JSON parsing error: " + err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.refreshClicked = false
			})
			return nil, nil
		}
	} else {
		log.Println("cannot fetch the flow page")
		toastText = "API error: cannot fetch the flow page"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.refreshClicked = false
		})
		return nil, nil
	}

	if resp.Code == 401 {
		toastText = "please log-in again"

		ctx.LocalStorage().Set("user", "")
		ctx.LocalStorage().Set("authGranted", false)
	}

	if len(resp.Posts) < 1 {
		toastText = "no posts to show; try adding some folks to your flow, or create a new post!"
		toastType = "info"
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.refreshClicked = false
		c.toastText = toastText
		c.toastType = toastType

		c.key = resp.Key
		if resp.Key != "" {
			c.user = c.users[resp.Key]
		}
	})

	return resp.Posts, resp.Users
}

type URIParts struct {
	SinglePost   bool
	SinglePostID string
	UserFlow     bool
	UserFlowNick string
	Hashtag      string
}

func (c *flowContent) parseFlowURI(ctx app.Context) URIParts {
	parts := URIParts{
		SinglePost:   false,
		SinglePostID: "",
		UserFlow:     false,
		UserFlowNick: "",
		Hashtag:      "",
	}

	url := strings.Split(ctx.Page().URL().Path, "/")

	if len(url) > 3 && url[3] != "" {
		switch url[2] {
		case "post":
			parts.SinglePost = true
			parts.SinglePostID = url[3]
			break

		case "user":
			parts.UserFlow = true
			parts.UserFlowNick = url[3]
			break

		case "hashtag":
			parts.Hashtag = url[3]
			break
		}
	}

	isPost := true
	if _, err := strconv.Atoi(parts.SinglePostID); parts.SinglePostID != "" && err != nil {
		// prolly not a post ID, but an user's nickname
		isPost = false
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.isPost = isPost
		c.userFlowNick = parts.UserFlowNick
		c.singlePostID = parts.SinglePostID
		c.hashtag = parts.Hashtag
	})

	return parts
}

func (c *flowContent) OnMount(ctx app.Context) {
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("image", c.handleImage)
	ctx.Handle("reply", c.handleReply)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("star", c.handleStar)
	ctx.Handle("dismiss", c.handleDismiss)
	ctx.Handle("refresh", c.handleRefresh)
	//ctx.Handle("message", c.handleNewPost)

	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1
	c.pageNoToFetch = 0
	c.lastPageFetched = false

	c.deletePostModalShow = false
	c.deleteModalButtonsDisabled = false

	c.eventListener = app.Window().AddEventListener("scroll", c.onScroll)
	//c.eventListenerMsg = app.Window().AddEventListener("message", c.onMessage)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
	//c.dismissEventListener = app.Window().AddEventListener("click", c.onClickGeneric)
}

func (c *flowContent) OnDismount() {
	// https://go-app.dev/reference#BrowserWindow
	//c.eventListener()
}

func (c *flowContent) OnNav(ctx app.Context) {
	ctx.Dispatch(func(ctx app.Context) {
		c.loaderShow = true
		c.loaderShowImage = true
		c.contentLoadFinished = false

		c.toastText = ""

		c.posts = nil
		c.users = nil
	})

	toastText := ""
	toastType := ""

	isPost := true

	ctx.Async(func() {
		parts := c.parseFlowURI(ctx)

		opts := pageOptions{
			PageNo:   0,
			Context:  ctx,
			CallerID: c.user.Nickname,

			SinglePost:   parts.SinglePost,
			SinglePostID: parts.SinglePostID,
			UserFlow:     parts.UserFlow,
			UserFlowNick: parts.UserFlowNick,
			Hashtag:      parts.Hashtag,
			HideReplies:  c.hideReplies,
		}

		posts, users := c.fetchFlowPage(opts)

		// try the singlePostID/userFlowNick var if present
		if parts.SinglePostID != "" && parts.SinglePost {
			if _, found := posts[parts.SinglePostID]; !found {
				toastText = "post not found"
			}
		}
		if parts.UserFlowNick != "" && parts.UserFlow {
			if _, found := users[parts.UserFlowNick]; !found {
				toastText = "user not found"
			}

			if value, found := c.user.FlowList[parts.UserFlowNick]; !value || !found {
				toastText = "follow the user to see their posts"
				toastType = "info"
			}
			isPost = false
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.pagination = 25
			c.pageNo = 1
			c.pageNoToFetch = 1

			c.user = users[c.key]

			c.users = users
			c.posts = posts
			c.singlePostID = parts.SinglePostID
			c.userFlowNick = parts.UserFlowNick
			c.isPost = isPost
			c.hashtag = parts.Hashtag

			if toastText != "" {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.toastType = toastType
			}

			c.loaderShow = false
			c.loaderShowImage = false
			c.contentLoadFinished = true
		})
		return
	})
}

func (c *flowContent) sortPosts() []models.Post {
	var sortedPosts []models.Post

	posts := c.posts
	if posts == nil {
		posts = make(map[string]models.Post)
	}

	flowList := c.user.FlowList
	if len(flowList) == 0 {
		return sortedPosts
	}

	// fetch posts and put them in an array
	for _, sortedPost := range posts {
		// do not append a post that is not meant to be shown
		if !c.user.FlowList[sortedPost.Nickname] && sortedPost.Nickname != "system" && sortedPost.Nickname != c.userFlowNick {
			continue
		}

		sortedPosts = append(sortedPosts, sortedPost)
	}

	return sortedPosts
}

func (c *flowContent) Render() app.UI {
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

	counter := 0

	sortedPosts := c.sortPosts()

	// order posts by timestamp DESC
	sort.SliceStable(sortedPosts, func(i, j int) bool {
		if c.singlePostID != "" {
			return sortedPosts[i].Timestamp.Before(sortedPosts[j].Timestamp)
		}

		return sortedPosts[i].Timestamp.After(sortedPosts[j].Timestamp)
	})

	// compose a summary of a long post to be replied to
	replySummary := ""
	if c.modalReplyActive && len(c.posts[c.interactedPostKey].Content) > configs.MaxPostLength {
		replySummary = c.posts[c.interactedPostKey].Content[:configs.MaxPostLength/10] + "- [...]"
	}

	return app.Main().Class("responsive").Body(
		// page heading
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.If(c.userFlowNick != "" && !c.isPost,
					app.H5().Body(
						app.Text(c.userFlowNick+"'s flow"),

						app.If(c.users[c.userFlowNick].Private,
							app.Span().Class("bold").Body(
								app.I().Text("lock"),
							),
						),
					),
				).ElseIf(c.singlePostID != "" && c.isPost,
					app.H5().Text("single post and replies"),
				).ElseIf(c.hashtag != "" && len(c.hashtag) < 20,
					app.H5().Text("hashtag #"+c.hashtag),
				).ElseIf(c.hashtag != "" && len(c.hashtag) >= 20,
					app.H5().Text("hashtag"),
				).Else(
					app.H5().Text("flow"),
					//app.P().Text("exclusive content incoming frfr"),
				),
			),

			app.Div().Class("small-padding").Body(
				app.Button().Title("refresh flow [R]").Class("border black white-text bold").Style("border-radius", "8px").OnClick(c.onClickRefresh).Disabled(c.postButtonsDisabled).Body(
					app.If(c.refreshClicked,
						app.Progress().Class("circle deep-orange-border small"),
					),
					app.Text("refresh"),
				),
			),
		),

		app.If(c.userFlowNick != "" && !c.isPost,
			app.Div().Class("row top-padding").Body(
				app.Img().Class("responsive max left").Src(c.users[c.userFlowNick].AvatarURL).Style("max-width", "80px").Style("border-radius", "50%"),
				/*;app.P().Class("max").Body(
					app.A().Class("bold deep-orange-text").Text(c.singlePostID).ID(c.singlePostID),
					//app.B().Text(post.Nickname).Class("deep-orange-text"),
				),*/

				//app.If(c.users[c.userFlowNick].About != "",
				app.Article().Class("max").Style("word-break", "break-word").Style("hyphens", "auto").Text(c.users[c.userFlowNick].About),
				//),
				app.Button().ID(c.userFlowNick).Class("black border white-text").Style("border-radius", "8px").OnClick(c.onClickFollow).Disabled(c.buttonDisabled || c.userFlowNick == c.user.Nickname).Body(
					app.If(c.user.FlowList[c.userFlowNick],
						app.Span().Text("unfollow"),
					).Else(
						app.Span().Text("follow"),
					),
				),
			),
		),

		app.Div().Class("space"),

		// snackbar
		app.If(c.toastText != "",
			app.A().OnClick(c.onClickDismiss).Body(
				app.Div().ID("snackbar").Class("snackbar white-text top active "+toastColor).Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// post deletion modal
		app.If(c.deletePostModalShow,
			app.Dialog().ID("delete-modal").Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("post deletion"),
				),

				app.Div().Class("space"),
				app.Article().Class("row").Body(
					app.I().Text("warning").Class("amber-text"),
					app.P().Class("max").Body(
						app.Span().Text("are you sure you want to delete your post?"),
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

		//app.Div().ID("overlay").Class("overlay").OnClick(c.onClickDismiss).Style("z-index", "50"),

		// sketchy reply modal
		app.If(c.modalReplyActive,
			app.Dialog().ID("reply-modal").Class("grey9 white-text center-align active").Style("max-width", "90%").Style("border-radius", "8px").Style("z-index", "75").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("reply"),
				),
				app.Div().Class("space"),

				app.Article().Class("post").Style("max-width", "100%").Body(
					app.If(replySummary != "",
						app.Details().Body(
							app.Summary().Text(replySummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
							app.Div().Class("space"),
							app.Span().Text(c.posts[c.interactedPostKey].Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic"),
						),
					).Else(
						app.Span().Text(c.posts[c.interactedPostKey].Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic"),
					),
				),

				app.Div().Class("field label textarea border extra deep-orange-text").Body(
					//app.Textarea().Class("active").Name("replyPost").OnChange(c.ValueTo(&c.replyPostContent)).AutoFocus(true).Placeholder("reply to: "+c.posts[c.interactedPostKey].Nickname),
					app.Textarea().Class("active").Name("replyPost").Text(c.replyPostContent).OnChange(c.ValueTo(&c.replyPostContent)).AutoFocus(true).ID("reply-textarea"),
					app.Label().Text("reply to: "+c.posts[c.interactedPostKey].Nickname).Class("active deep-orange-text"),
					//app.Label().Text("text").Class("active"),
				),
				app.Div().Class("field label border extra deep-orange-text").Body(
					app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload).Accept("image/*"),
					app.Input().Class("active").Type("text").Value(c.newFigFile).Disabled(true),
					app.Label().Text("image").Class("active deep-orange-text"),
					app.I().Text("image"),
				),

				app.Div().Class("row").Body(
					app.Button().Class("max border deep-orange7 white-text bold").Text("cancel").Style("border-radius", "8px").OnClick(c.onClickDismiss).Disabled(c.postButtonsDisabled),
					app.Button().ID("reply").Class("max border deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickPostReply).Disabled(c.postButtonsDisabled).Body(
						app.If(c.postButtonsDisabled,
							app.Progress().Class("circle white-border small"),
						),
						app.Text("reply"),
					),
				),
				app.Div().Class("space"),
			),
		),

		// flow posts/articles
		app.Table().Class("left-aligni border").ID("table-flow").Style("padding", "0 0 2em 0").Style("border-spacing", "0.1em").Body(
			// table body
			app.TBody().Body(
				//app.Range(c.posts).Map(func(key string) app.UI {
				//app.Range(pagedPosts).Slice(func(idx int) app.UI {
				app.Range(sortedPosts).Slice(func(idx int) app.UI {
					counter++
					if counter > c.pagination*c.pageNo {
						return nil
					}

					//post := c.sortedPosts[idx]
					post := sortedPosts[idx]
					key := post.ID

					previousContent := ""

					// prepare reply parameters to render
					if post.ReplyToID != "" {
						if c.hideReplies {
							return nil
						}

						if previous, found := c.posts[post.ReplyToID]; found {
							if value, foundU := c.user.FlowList[previous.Nickname]; (!value || !foundU) && c.users[previous.Nickname].Private {
								previousContent = "this content is private"
							} else {
								previousContent = previous.Nickname + " posted: " + previous.Content
							}
						} else {
							previousContent = "the post was deleted bye"
						}
					}

					// filter out not-single-post items
					if c.singlePostID != "" {
						if c.isPost && post.ID != c.singlePostID && c.singlePostID != post.ReplyToID {
							return nil
						}

						if _, found := c.users[c.singlePostID]; (!c.isPost && !found) || (found && post.Nickname != c.singlePostID) {
							return nil
						}
					}

					if c.userFlowNick != "" {
						if post.Nickname != c.userFlowNick {
							return nil
						}
					}

					// only show posts of users in one's flowList
					if !c.user.FlowList[post.Nickname] && post.Nickname != "system" {
						return nil
					}

					// check the post's length, on threshold use <details> tag
					postDetailsSummary := ""
					if len(post.Content) > configs.MaxPostLength {
						postDetailsSummary = post.Content[:configs.MaxPostLength/10] + "- [...]"
					}

					// the same as above with the previous post's length for reply render
					previousDetailsSummary := ""
					if len(previousContent) > configs.MaxPostLength {
						previousDetailsSummary = previousContent[:configs.MaxPostLength/10] + "- [...]"
					}

					// fetch the image
					var imgSrc string

					// check the URL/URI format
					if post.Type == "fig" {
						if _, err := url.ParseRequestURI(post.Content); err == nil {
							imgSrc = post.Content
						} else {
							fileExplode := strings.Split(post.Content, ".")
							extension := fileExplode[len(fileExplode)-1]

							imgSrc = "/web/pix/thumb_" + post.Content
							if extension == "gif" {
								imgSrc = "/web/click-to-see-gif.jpg"
							}
						}
					} else if post.Type == "post" {
						if _, err := url.ParseRequestURI(post.Figure); err == nil {
							imgSrc = post.Figure
						} else {
							fileExplode := strings.Split(post.Figure, ".")
							extension := fileExplode[len(fileExplode)-1]

							imgSrc = "/web/pix/thumb_" + post.Figure
							if extension == "gif" {
								imgSrc = "/web/click-to-see.gif"
							}
						}
					}

					// fetch binary image data
					/*if post.Type == "fig" && imgSrc == "" {
						payload := struct {
							PostID  string `json:"post_id"`
							Content string `json:"content"`
						}{
							PostID:  post.ID,
							Content: post.Content,
						}

						var resp *[]byte
						var ok bool

						if resp, ok = littrAPI("POST", "/api/pix", payload, c.user.Nickname); !ok {
							log.Println("api failed")
							imgSrc = "/web/android-chrome-512x512.png"
						} else {
							imgSrc = "data:image/*;base64," + b64.StdEncoding.EncodeToString(*resp)
						}
					}*/

					var postTimestamp string

					// use JS toLocaleString() function to reformat the timestamp
					// use negated LocalTimeMode boolean as true! (HELP)
					if !c.user.LocalTimeMode {
						postLocale := app.Window().
							Get("Date").
							New(post.Timestamp.Format(time.RFC3339))

						postTimestamp = postLocale.Call("toLocaleString", "en-GB").String()
					} else {
						postTimestamp = post.Timestamp.Format("Jan 02, 2006 / 15:04:05")
					}

					// omit older system messages for new users
					if post.Nickname == "system" && post.Timestamp.Before(c.user.RegisteredTime) {
						return nil
					}

					systemLink := "/polls"
					if post.Nickname == "system" && post.Type == "user" {
						systemLink = "/flow/user/" + post.Figure
					}

					return app.Tr().Class().Class("bottom-padding").Body(
						// special system post
						app.If(post.Nickname == "system",
							app.Td().Class("post align-left").Attr("touch-action", "none").Body(
								app.Article().Class("responsive margin-top center-align").Body(
									app.A().Href(systemLink).Body(
										app.Span().Class("bold").Text(post.Content),
									),
								),
								app.Div().Class("row").Body(
									app.Div().Class("max").Body(
										//app.Text(post.Timestamp.Format("Jan 02, 2006 / 15:04:05")),
										app.Text(postTimestamp),
									),
								),
							),

						// other posts
						).Else(
							//app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).On("scroll", c.onScroll).Body(
							app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).Attr("touch-action", "none").Body(

								// post header (author avatar + name + link button)
								app.Div().Class("row top-padding").Body(
									app.Img().Title("user's avatar").Class("responsive max left").Src(c.users[post.Nickname].AvatarURL).Style("max-width", "60px").Style("border-radius", "50%"),
									app.P().Class("max").Body(
										app.A().Title("user's flow link").Class("bold deep-orange-text").OnClick(c.onClickUserFlow).ID(post.Nickname).Body(
											app.Span().Class("large-text bold deep-orange-text").Text(post.Nickname),
										),
										//app.B().Text(post.Nickname).Class("deep-orange-text"),
									),
									app.Button().ID(key).Title("link to this post (to clipboard)").Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
										app.I().Text("link"),
									),
								),

								// pic post
								app.If(post.Type == "fig",
									app.Article().Style("z-index", "5").Style("border-radius", "8px").Class("transparent medium no-margin").Body(
										app.If(c.loaderShowImage,
											app.Div().Class("small-space"),
											app.Div().Class("loader center large deep-orange active"),
										),
										//app.Img().Class("no-padding center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
										app.Img().Class("no-padding center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
									),

								// reply + post
								).Else(
									app.If(post.ReplyToID != "",
										app.Article().Class("black-text yellow10").Style("border-radius", "8px").Style("max-width", "100%").Body(
											app.Div().Class("row max").Body(
												app.If(previousDetailsSummary != "",
													app.Details().Class("max").Body(
														app.Summary().Text(previousDetailsSummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
														app.Div().Class("space"),
														app.Span().Class("bold").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
													),
												).Else(
													app.Span().Class("max bold").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
												),

												app.Button().Title("link to original post").ID(post.ReplyToID).Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
													app.I().Text("history"),
												),
											),
										),
									),

									app.If(len(post.Content) > 0,
										app.Article().Class("surface-container-highest").Style("border-radius", "8px").Style("max-width", "100%").Body(
											app.If(postDetailsSummary != "",
												app.Details().Body(
													app.Summary().Text(postDetailsSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
													app.Div().Class("space"),
													app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
												),
											).Else(
												app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
											),
										),
									),

									app.If(post.Figure != "",
										app.Article().Style("z-index", "4").Style("border-radius", "8px").Class("transparent medium medium").Body(
											app.If(c.loaderShowImage,
												app.Div().Class("small-space"),
												app.Div().Class("loader center large deep-orange active"),
											),
											//app.Img().Class("no-padding center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
											app.Img().Class("no-padding center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
										),
									),
								),

								// post footer (timestamp + reply buttom + star/delete button)
								app.Div().Class("row").Body(
									app.Div().Class("max").Body(
										//app.Text(post.Timestamp.Format("Jan 02, 2006 / 15:04:05")),
										app.Text(postTimestamp),
									),
									app.If(post.Nickname != "system",
										app.If(post.ReplyCount > 0,
											app.B().Title("reply count").Text(post.ReplyCount).Class("left-padding"),
										),
										app.Button().Title("reply").ID(key).Class("transparent circle").OnClick(c.onClickReply).Disabled(c.buttonDisabled).Body(
											app.I().Text("reply"),
										),
									),
									app.If(c.user.Nickname == post.Nickname,
										app.B().Title("reaction count").Text(post.ReactionCount).Class("left-padding"),
										//app.Button().ID(key).Class("transparent circle").OnClick(c.onClickDelete).Disabled(c.buttonDisabled).Body(
										app.Button().Title("delete this post").ID(key).Class("transparent circle").OnClick(c.onClickDeleteButton).Disabled(c.buttonDisabled).Body(
											app.I().Text("delete"),
										),
									).Else(
										app.B().Title("reaction count").Text(post.ReactionCount).Class("left-padding"),
										app.Button().Title("increase the reaction count").ID(key).Class("transparent circle").OnClick(c.onClickStar).Disabled(c.buttonDisabled).Attr("touch-action", "none").Body(
											//app.I().Text("ac_unit"),
											app.I().Text("bomb"),
										),
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
