package users

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

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

	toast := common.Toast{AppContext: &ctx}

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

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		// delete the request from one's requestList
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		//toast.Text("requested removed").Type("success").Dispatch(c, dispatch)

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

		input2 := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + nick + "/lists",
			Data:        payload2,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output2 := &common.Response{}

		if ok := common.FetchData(input2, output2); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output2.Code != 200 {
			toast.Text(output2.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		toast.Text(common.MSG_USER_UPDATED_SUCCESS).Type(common.TTYPE_INFO).Dispatch(c, dispatch)

		/*ctx.Dispatch(func(ctx app.Context) {
			//c.user.FlowList = ourFlowList
			//c.users[c.user.Nickname] = payload
		})*/
		return
	})
	return
}

func (c *Content) onClickCancel(ctx app.Context, e app.Event) {
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	toast := common.Toast{AppContext: &ctx}

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

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		// delete the request from one's requestList
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		toast.Text(common.MSG_FOLLOW_REQUEST_REMOVED).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
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

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		user := c.users[nick]

		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/users/" + nick + "/request",
			Data:        nil,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}
		user.RequestList[c.user.Nickname] = false

		toast.Text(common.MSG_USER_FOLLOW_REQ_REMOVED).Type(common.TTYPE_INFO).Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
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

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		user := c.users[nick]

		input := &common.CallInput{
			Method:      "POST",
			Url:         "/api/v1/users/" + nick + "/request",
			Data:        nil,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}
		user.RequestList[c.user.Nickname] = true

		toast.Text(common.MSG_REQ_TO_FOLLOW_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			c.users[nick] = user
		})
		return
	})
	return
}

func (c *Content) onClickUser(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("preview", key)

	ctx.Dispatch(func(ctx app.Context) {
		c.usersButtonDisabled = true
		c.showUserPreviewModal = true
	})
	return
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

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		payload := struct {
			FlowList  map[string]bool `json:"flow_list"`
			ShadeList map[string]bool `json:"shade_list"`
		}{
			FlowList:  userShaded.FlowList,
			ShadeList: userShaded.ShadeList,
		}

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + userShaded.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
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

		input = &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output = &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 && output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		ctx.LocalStorage().Set("user", c.user)
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
