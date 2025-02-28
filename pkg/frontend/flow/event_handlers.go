package flow

import (
	"log"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) onClickFollow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	if key == c.user.Nickname {
		return
	}

	user := c.user
	flowList := user.FlowList

	if c.user.ShadeList[key] {
		return
	}

	if flowList == nil {
		flowList = make(map[string]bool)
		flowList[c.user.Nickname] = true
	}

	toast := common.Toast{AppContext: &ctx}

	if value, found := flowList[key]; found {
		if !value && c.users[key].Private {
			toast.Text(common.ERR_PRIVATE_ACC).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				c.buttonDisabled = false
				c.postButtonsDisabled = false
			})
			return
		}
		flowList[key] = !flowList[key]
	} else {
		if c.users[key].Private {
			toast.Text(common.ERR_PRIVATE_ACC).Type(common.TTYPE_ERR).Dispatch()

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

		defer ctx.Dispatch(func(ctx app.Context) {
			c.buttonDisabled = false
			c.postButtonsDisabled = false
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
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			flowList[key] = !flowList[key]
			return
		}

		if output.Code != 200 && output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			flowList[key] = !flowList[key]
			return
		}

		user.FlowList = flowList
		common.SaveUser(&user, &ctx)

		ctx.Dispatch(func(ctx app.Context) {
			c.user = user
			c.users[user.Nickname] = user
		})

		ctx.NewAction("refresh")
	})
}

/*func (c *Content) onClickDismiss(ctx app.Context, e app.Event) {
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
	}
}*/

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

// From ModalPostDelete modal...
func (c *Content) onClickDelete(ctx app.Context, e app.Event) {
	key := c.interactedPostKey
	ctx.NewActionWithValue("delete", key)

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteModalButtonsDisabled = true
	})
}

/*func (c *Content) onMessage(ctx app.Context, e app.Event) {
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
}*/
