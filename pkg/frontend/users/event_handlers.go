package users

import (
	"log"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type stub struct{}

func (c *Content) onClick(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("toggle", key)
	c.usersButtonDisabled = true
}

func (c *Content) onClickAllow(ctx app.Context, e app.Event) {
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	toastText := ""
	toastType := "error"

	ctx.Async(func() {
		user := c.user

		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}

		user.RequestList[nick] = false

		payload := struct {
			RequestList map[string]bool `json:"request_list"`
		}{
			RequestList: user.RequestList,
		}

		input := common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// delete the request from one's requestList
		if ok := common.CallAPI(input, &stub{}); !ok {
			toastText = "problem calling the backend"
			toastType = "error"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.toastType = toastType
				c.usersButtonDisabled = false
			})
			return
		} else {
			toastText = "requested removed"
			toastType = "success"
		}

		// prepare the lists for the counterpart
		fellowFlowList := make(map[string]bool)
		fellowFlowList[nick] = true
		fellowFlowList[c.user.Nickname] = true

		//ourFlowList := c.user.FlowList
		//ourFlowList[nick] = true

		payload2 := struct {
			FlowList map[string]bool `json:"flow_list"`
		}{
			FlowList: fellowFlowList,
		}

		input2 := common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + nick + "/lists",
			Data:        payload2,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		if ok := common.CallAPI(input2, &payload2); !ok {
			toastText = "problem calling the backend"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
				c.toastType = toastType
				c.usersButtonDisabled = false
			})
			return
		} else {
			toastText = "user updated, request removed"
			toastType = "success"
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.toastType = toastType
			c.usersButtonDisabled = false

			//c.user.FlowList = ourFlowList
			//c.users[c.user.Nickname] = payload
		})
		return
	})
	return
}

func (c *Content) onClickCancel(ctx app.Context, e app.Event) {
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	toastText := ""

	ctx.Async(func() {
		user := c.user
		toastType := "error"

		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}

		user.RequestList[nick] = false

		payload := struct {
			RequestList map[string]bool `json:"request_list"`
		}{
			RequestList: user.RequestList,
		}

		input := common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// delete the request from one's requestList
		if ok := common.CallAPI(input, &stub{}); !ok {
			toastText = "problem calling the backend"
			toastType = "error"
		} else {
			toastText = "requested removed"
			toastType = "success"
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.toastType = toastType
			c.usersButtonDisabled = false

			c.user = user
			c.users[c.user.Nickname] = user
		})
		return
	})
	return
}

func (c *Content) onClickPrivateOff(ctx app.Context, e app.Event) {
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	toastText := ""

	ctx.Async(func() {
		user := c.users[nick]
		toastType := "error"

		input := common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/users/" + nick + "/request",
			Data:        nil,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		if ok := common.CallAPI(input, &stub{}); !ok {
			toastText = "problem calling the backend"
		} else {
			toastText = "request to follow removed"
			toastType = "info"

			if user.RequestList == nil {
				user.RequestList = make(map[string]bool)
			}
			user.RequestList[c.user.Nickname] = false
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.toastType = toastType
			c.usersButtonDisabled = false

			c.users[nick] = user
		})
		return
	})
	return
}

func (c *Content) onClickPrivateOn(ctx app.Context, e app.Event) {
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	toastText := ""

	ctx.Async(func() {
		user := c.users[nick]
		toastType := "error"

		input := common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/users/" + nick + "/request",
			Data:        nil,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		if ok := common.CallAPI(input, &stub{}); !ok {
			toastText = "problem calling the backend"
		} else {
			toastText = "requested to follow"
			toastType = "success"

			if user.RequestList == nil {
				user.RequestList = make(map[string]bool)
			}
			user.RequestList[c.user.Nickname] = true
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.toastType = toastType
			c.usersButtonDisabled = false

			c.users[nick] = user
		})
		return
	})
	return
}

func (c *Content) onClickUser(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("preview", key)
	c.usersButtonDisabled = true
	c.showUserPreviewModal = true
}

func (c *Content) onClickUserFlow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	c.usersButtonDisabled = true

	// isn't the use blocked?
	if c.user.ShadeList[key] {
		c.usersButtonDisabled = false
		return
	}

	// is the user followed?
	if !c.user.FlowList[key] {
		c.usersButtonDisabled = false
		return
	}

	// show only 1+ posts
	if c.userStats[key].PostCount == 0 {
		c.usersButtonDisabled = false
		return
	}

	ctx.Navigate("/flow/user/" + key)
}

func (c *Content) onClickUserShade(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	c.usersButtonDisabled = true

	// do not shade yourself
	if c.user.Nickname == key {
		c.usersButtonDisabled = false
		return
	}

	// fetch the to-be-shaded user
	userShaded, found := c.users[key]
	if !found {
		c.usersButtonDisabled = false
		return
	}

	if userShaded.FlowList == nil {
		userShaded.FlowList = make(map[string]bool)
	}

	// disable any following of such user
	userShaded.FlowList[c.user.Nickname] = false
	c.user.FlowList[key] = false

	// negate the previous state
	shadeListItem := c.user.ShadeList[key]

	if c.user.ShadeList == nil {
		c.user.ShadeList = make(map[string]bool)
	}

	if key != c.user.Nickname {
		c.user.ShadeList[key] = !shadeListItem
	}

	toastText := ""

	ctx.Async(func() {
		payload := struct {
			FlowList  map[string]bool `json:"flow_list"`
			ShadeList map[string]bool `json:"shade_list"`
		}{
			FlowList:  userShaded.FlowList,
			ShadeList: userShaded.ShadeList,
		}

		input := common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + userShaded.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if ok := common.CallAPI(input, &response); !ok {
			toastText = "generic backend error"
			return
		}

		if response.Code != 200 && response.Code != 201 {
			toastText = "user update failed: " + response.Message
			log.Println(response.Message)
			return
		}

		/*var stream []byte
		if err := reload(c.user, &stream); err != nil {
			toastText = "local storage reload failed: " + err.Error()
			return
		}*/

		payload = struct {
			FlowList  map[string]bool `json:"flow_list"`
			ShadeList map[string]bool `json:"shade_list"`
		}{
			FlowList:  c.user.FlowList,
			ShadeList: c.user.ShadeList,
		}

		input = common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		if ok := common.CallAPI(input, &response); !ok {
			toastText = "generic backend error"
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			//ctx.LocalStorage().Set("user", stream)

			c.toastText = toastText
			c.toastShow = (toastText != "")
			c.usersButtonDisabled = false

			log.Println("dispatch ends")
		})
	})

	c.userButtonDisabled = false
	return

}

func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}
}

func (c *Content) onScroll(ctx app.Context, e app.Event) {
	ctx.NewAction("scroll")
}

func (c *Content) onSearch(ctx app.Context, e app.Event) {
	val := ctx.JSSrc().Get("value").String()

	if len(val) > 20 {
		return
	}

	ctx.NewActionWithValue("search", val)
}
