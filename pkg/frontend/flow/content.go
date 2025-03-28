// The very flow view and view-controllers logic package.
package flow

import (
	"encoding/base64"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	loaderShow      bool
	loaderShowImage bool

	hideReplies bool

	contentLoadFinished bool

	user models.User
	key  string

	toast common.Toast

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

	postKey string
	posts   map[string]models.Post
	users   map[string]models.User

	refreshClicked bool

	hashtag string
}

func (c *Content) OnMount(ctx app.Context) {
	if app.IsServer {
		return
	}

	ctx.Handle("ask", c.handlePrivateMode)
	ctx.Handle("blur-post", c.handleTextareaBlur)
	ctx.Handle("cancel", c.handlePrivateMode)
	ctx.Handle("clear", c.handleClear)
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("dismiss", c.handleDismiss)
	ctx.Handle("follow", c.handleToggle)
	ctx.Handle("history", c.handleLink)
	ctx.Handle("image-click", c.handleImage)
	ctx.Handle("link", c.handleLink)
	ctx.Handle("modal-post-delete", c.handleModalPostDeleteShow)
	ctx.Handle("modal-post-reply", c.handleModalPostReplyShow)
	ctx.Handle("mouse-enter", c.handleMouseEnter)
	ctx.Handle("mouse-leave", c.handleMouseLeave)
	ctx.Handle("refresh", c.handleRefresh)
	ctx.Handle("reply", c.handleReply)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("shade", c.handleUserShade)
	ctx.Handle("star", c.handleStar)
	ctx.Handle("unfollow", c.handleToggle)
	ctx.Handle("user", c.handleUser)

	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1
	c.pageNoToFetch = 0
	c.lastPageFetched = false

	c.deletePostModalShow = false
	c.deleteModalButtonsDisabled = false

	ctx.GetState(common.StateNameUser, &c.user)

	// Load the saved draft from localStorage.
	_ = ctx.LocalStorage().Get("newReplyDraft", &c.replyPostContent)
	_ = ctx.LocalStorage().Get("newReplyFigFile", &c.newFigFile)

	var data string
	_ = ctx.LocalStorage().Get("newReplyFigData", &data)

	c.newFigData, _ = base64.StdEncoding.DecodeString(data)
}

func (c *Content) OnNav(ctx app.Context) {
	if app.IsServer {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.loaderShow = true
		c.loaderShowImage = true
		c.contentLoadFinished = false

		c.toast.TText = ""

		c.posts = nil
		c.users = nil
	})

	toast := common.Toast{AppContext: &ctx}

	isPost := true

	ctx.Async(func() {
		parts := c.parseFlowURI(ctx)

		opts := pageOptions{
			PageNo:  0,
			Context: ctx,

			SinglePost:   parts.SinglePost,
			SinglePostID: parts.SinglePostID,
			UserFlow:     parts.UserFlow,
			UserFlowNick: parts.UserFlowNick,
			Hashtag:      parts.Hashtag,
			HideReplies:  c.hideReplies,
		}

		posts, users := c.fetchFlowPage(opts)

		// The content to render is to show the singlePost view.
		if parts.SinglePostID != "" && parts.SinglePost && posts != nil {
			if _, found := (*posts)[parts.SinglePostID]; !found {
				toast.Text(common.ERR_POST_NOT_FOUND).Type(common.TTYPE_ERR).Link("/flow").Dispatch()
			}
		}

		// The content to render is to show the userFlow view.
		if parts.UserFlowNick != "" && parts.UserFlow && users != nil {
			if _, found := (*users)[parts.UserFlowNick]; !found {
				toast.Text(common.ERR_USER_NOT_FOUND).Type(common.TTYPE_ERR).Link("/flow").Dispatch()
			}

			isPost = false
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.pagination = 25
			c.pageNo = 1
			c.pageNoToFetch = 1

			if users != nil {
				c.user = (*users)[c.key]
				c.users = *users

				// Also update the user struct in the LS. Don't catch error as other vars are loaded into the component.
				ctx.SetState(common.StateNameUser, c.user).Persist()
			}

			if posts != nil {
				c.posts = *posts
			}

			c.singlePostID = parts.SinglePostID
			c.userFlowNick = parts.UserFlowNick
			c.isPost = isPost
			c.hashtag = parts.Hashtag

			c.loaderShow = false
			c.loaderShowImage = false
			c.contentLoadFinished = true
		})
	})
}
