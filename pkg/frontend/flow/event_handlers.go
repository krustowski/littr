package flow

import (
	"log"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) onClickFollow(ctx app.Context, e app.Event) {
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

	toast := common.Toast{AppContext: &ctx}

	if value, found := flowList[key]; found {
		if !value && c.users[key].Private {
			toast.Text(common.ERR_PRIVATE_ACC).Type(common.TTYPE_ERR).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonDisabled = false
				c.postButtonsDisabled = false
			})
			return
		}
		flowList[key] = !flowList[key]
	} else {
		if c.users[key].Private {
			toast.Text(common.ERR_PRIVATE_ACC).Type(common.TTYPE_ERR).Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonDisabled = false
				c.postButtonsDisabled = false
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

		payload := struct {
			FlowList map[string]bool `json:"flow_list"`
		}{
			FlowList: flowList,
		}

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: c.hideReplies,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 && output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.buttonDisabled = false
			c.postButtonsDisabled = false

			c.user.FlowList = flowList
		})
	})
}

func (c *Content) onClickLink(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host

	// write the link to browsers's clipboard
	navigator := app.Window().Get("navigator")
	if !navigator.IsNull() {
		clipboard := navigator.Get("clipboard")
		if !clipboard.IsNull() && !clipboard.IsUndefined() {
			clipboard.Call("writeText", scheme+"://"+host+"/flow/posts/"+key)
		}
	}
	ctx.Navigate("/flow/posts/" + key)
}

func (c *Content) onClickDismiss(ctx app.Context, e app.Event) {
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

func (c *Content) onClickUserFlow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	//c.buttonDisabled = true

	ctx.Navigate("/flow/user/" + key)
}

// onClickReply acts like a caller function evoked when user click on the reply icon at one's post
func (c *Content) onClickReply(ctx app.Context, e app.Event) {
	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPostKey = ctx.JSSrc().Get("id").String()
		c.modalReplyActive = true
		c.postButtonsDisabled = false
		c.buttonDisabled = true
	})
}

// onClickPostReply acts like a caller function evoked when user clicks on "reply" button in the reply modal
func (c *Content) onClickPostReply(ctx app.Context, e app.Event) {
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

func (c *Content) onClickGeneric(ctx app.Context, e app.Event) {
	if e.Get("target").String() == "overlay" || e.Get("srcElement").String() == "overlay" {
		log.Println("overlay clicked")
	}
}

func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
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

		ctx.NewAction("dismiss")
		ctx.NewAction("clear")
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

func (c *Content) onClickImage(ctx app.Context, e app.Event) {
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
			toast.Text(err.Error()).Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = false
			})
			return

		} else {
			toast.Text("image is ready").Type("info").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.postButtonsDisabled = false

				c.newFigFile = file.Get("name").String()
				c.newFigData = data
			})
			return

		}
	})
}

func (c *Content) onClickDeleteButton(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPostKey = key
		c.deleteModalButtonsDisabled = false
		c.deletePostModalShow = true
	})
}

func (c *Content) onClickDelete(ctx app.Context, e app.Event) {
	key := c.interactedPostKey
	ctx.NewActionWithValue("delete", key)

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteModalButtonsDisabled = true
	})
}

func (c *Content) onScroll(ctx app.Context, e app.Event) {
	ctx.NewAction("scroll")
}

func (c *Content) onClickRefresh(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
	ctx.NewAction("clear")
	ctx.NewAction("refresh")
}

func (c *Content) onMessage(ctx app.Context, e app.Event) {
	data := e.JSValue().Get("data").String()
	log.Println("msg event: data:" + data)

	if data == "heartbeat" || data == c.user.Nickname {
		return
	}

	if _, flowed := c.user.FlowList[data]; !flowed {
		return
	}

	toast := common.Toast{AppContext: &ctx}
	toast.Text("new post added above").Type(common.TTYPE_INFO).Dispatch(c, dispatch)
}

func (c *Content) onClickStar(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("star", key)
}
