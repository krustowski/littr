package users

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// onClick prolly acts like a callback function for the generic user flow's state ((un)flowed toggling).
func (c *Content) onClick(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	c.usersButtonDisabled = true
	ctx.NewActionWithValue("toggle", key)
}

// onClickAllow is a callback function that enables the controlling user to accept the incoming follow request (one must be in the private mode of their profile).
func (c *Content) onClickAllow(ctx app.Context, e app.Event) {
	// Fetch the counterpart's nickname.
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		currentUser := c.user

		// Pathc the nil requestMap.
		if currentUser.RequestList == nil {
			currentUser.RequestList = make(map[string]bool)
		}

		// Falsify the incoming nick's request making it allowed soon.
		currentUser.RequestList[nick] = false

		// Prepare the request data structure.
		payload := struct {
			RequestList map[string]bool `json:"request_list"`
		}{
			RequestList: currentUser.RequestList,
		}

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API call response object.
		output := &common.Response{}

		// Delete the request from one's (the controlling one's) requestList.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Check the HTTP 200 response code, otherwise print the API response message in the toast.
		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Cast the successful API response in the toast.
		//toast.Text("requested removed").Type("success").Dispatch(c, dispatch)

		// Prepare the lists for the counterpart. Ensure that the controlling user's account is following, as well as their own one too.
		fellowFlowList := make(map[string]bool)
		fellowFlowList[nick] = true
		fellowFlowList[c.user.Nickname] = true

		// Update the controlling user's flowList (disabled, as it is not sure, whether such user wants to follow the counterpart directly too.
		//ourFlowList := c.user.FlowList
		//ourFlowList[nick] = true

		// Prepare the second request data structure.
		payload2 := struct {
			FlowList map[string]bool `json:"flow_list"`
		}{
			FlowList: fellowFlowList,
		}

		// Compose the second API input payload.
		input2 := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + nick + "/lists",
			Data:        payload2,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output2 := &common.Response{}

		// Update the counterpart's flowList.
		if ok := common.FetchData(input2, output2); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Check the HTTP 200 response code, otherwise print the API response message in the toast.
		if output2.Code != 200 {
			toast.Text(output2.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Cast the successful update of both lists.
		toast.Text(common.MSG_USER_UPDATED_SUCCESS).Type(common.TTYPE_INFO).Dispatch()

		common.SaveUser(&currentUser, &ctx)

		// Dispatch the updated controlling one's flowList.
		ctx.Dispatch(func(ctx app.Context) {
			c.user = currentUser
			c.users[c.user.Nickname] = currentUser
		})
	})
}

// onClickCancel is a callback function to cancel the incoming follow request.
func (c *Content) onClickCancel(ctx app.Context, e app.Event) {
	// Fetch the counterpart's nickname.
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		currentUser := c.user

		// Patch the requestList nil map.
		if currentUser.RequestList == nil {
			currentUser.RequestList = make(map[string]bool)
		}

		// Falsify the counterpart's existence in the controlling one's requestList (disable the request).
		currentUser.RequestList[nick] = false

		// Prepare the request data structure.
		payload := struct {
			RequestList map[string]bool `json:"request_list"`
		}{
			RequestList: currentUser.RequestList,
		}

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + currentUser.Nickname + "/lists",
			Data:        payload,
			CallerID:    currentUser.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output := &common.Response{}

		// Delete the request from the controlling one's requestList.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
		}

		// Check for the HTTP 200 response code, otherwise print the API response message in the toast.
		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Cast the successful request removal.
		toast.Text(common.MSG_FOLLOW_REQUEST_REMOVED).Type(common.TTYPE_SUCCESS).Dispatch()

		common.SaveUser(&currentUser, &ctx)

		// Dispatch the changes to match them in the UI.
		ctx.Dispatch(func(ctx app.Context) {
			c.user = currentUser
			c.users[c.user.Nickname] = currentUser
		})
	})
}

// onClickPrivateOff is to take back the previously sent follow request to the counterpart.
func (c *Content) onClickPrivateOff(ctx app.Context, e app.Event) {
	// Fetch the counterpart's nickname.
	nick := ctx.JSSrc().Get("id").String()
	if nick == "" {
		return
	}

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		user := c.users[nick]

		// Patch the nil requestList map.
		if user.RequestList == nil {
			user.RequestList = make(map[string]bool)
		}
		user.RequestList[c.user.Nickname] = false

		// Prepare the request data structure.
		payload := struct {
			RequestList map[string]bool `json:"request_list"`
		}{
			RequestList: user.RequestList,
		}

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + nick + "/lists",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output := &common.Response{}

		// Call the API to delete the follow request.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
		}

		// Check for the HTTP 200 response code, otherwise print the API response message in the toast.
		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Cast the successful request removal.
		toast.Text(common.MSG_FOLLOW_REQUEST_REMOVED).Type(common.TTYPE_INFO).Dispatch()

		// Dispatch the changes to match the reality in the UI.
		ctx.Dispatch(func(ctx app.Context) {
			c.users[nick] = user
		})
		return
	})
	return
}

// onClickPrivateOn is a callback function to send a follow request to an account that is in the private mode.
func (c *Content) onClickPrivateOn(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	if key == "" {
		return
	}

	// The counterpart.
	user := c.users[key]
	requestList := user.RequestList

	// Instantiate new toast struct.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// Patch the nil requestList map.
		if requestList == nil {
			requestList = make(map[string]bool)
		}
		requestList[c.user.Nickname] = true

		// Prepare the request data structure.
		payload := struct {
			RequestList map[string]bool `json:"request_list"`
		}{
			RequestList: requestList,
		}

		// Compose the API input payload.
		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + key + "/lists",
			Data:        payload,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response object.
		output := &common.Response{}

		// Patch the requestList of the counterpart user's.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			requestList[c.user.Nickname] = false

			ctx.Dispatch(func(ctx app.Context) {
				user.RequestList = requestList
				c.users[user.Nickname] = user
			})
			return
		}

		// Check for the HTTP 200 response code, otherwise print the API response message in the toast.
		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			requestList[c.user.Nickname] = false

			ctx.Dispatch(func(ctx app.Context) {
				user.RequestList = requestList
				c.users[user.Nickname] = user
			})
			return
		}

		// Cast the successful patch made.
		toast.Text(common.MSG_REQ_TO_FOLLOW_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch()

		// Dispatch the changes to match the reality in the UI.
		user.RequestList = requestList
		user.Searched = true

		ctx.Dispatch(func(ctx app.Context) {
			c.users[user.Nickname] = user
		})
		return
	})
	return
}

// onClickUser is a callback function that is called when one clicks on the user's nickname to show their user modal.
func (c *Content) onClickUser(ctx app.Context, e app.Event) {
	// Fetch the requested ID (nickname).
	key := ctx.JSSrc().Get("id").String()

	// Cast a new action with the key value.
	ctx.NewActionWithValue("preview", key)

	// Dispatch the UI changes to match the processing time.
	ctx.Dispatch(func(ctx app.Context) {
		c.usersButtonDisabled = true
		c.showUserPreviewModal = true
	})
	return
}

// onClickUserFlow is a callback function that enables one to be navigated to the counterpart user's flow (if permitted).
func (c *Content) onClickUserFlow(ctx app.Context, e app.Event) {
	// Fetch the requested ID (nickname).
	key := ctx.JSSrc().Get("id").String()

	// Nasty way of how to disable duttons (use Dispatch function instead + new action casting).
	c.usersButtonDisabled = true

	defer ctx.Dispatch(func(ctx app.Context) {
		c.usersButtonDisabled = false
	})

	if key == "" {
		return
	}

	// Check if the user is not block already.
	/*if c.user.ShadeList[key] {
		c.usersButtonDisabled = false
		return
	}

	// Check if the counterpart user is blocking us.
	if c.users[key].ShadeList[c.user.Nickname] {
		c.usersButtonDisabled = false
		return
	}

	// Check the state of the counterpart user's in the controlling one's flowList. If not followed, do not redirect anywhere.
	if !c.user.FlowList[key] {
		c.usersButtonDisabled = false
		return
	}*/

	// Check for the post count of the requested user. Only redirect to the flow of 1 and more posts.
	/*if c.userStats[key].PostCount == 0 {
		c.usersButtonDisabled = false
		return
	}*/

	// Navigate to the counterpart user's flow.
	ctx.Navigate("/flow/users/" + key)
}

// oncLickUserShade is a callback function that enables shadeList toggling.
func (c *Content) onClickUserShade(ctx app.Context, e app.Event) {
	// Fetch the requested ID (nickname).
	key := ctx.JSSrc().Get("id").String()

	// Nasty way of buttons disabling (use Dispatch function + new action casting).
	ctx.Dispatch(func(ctx app.Context) {
		c.usersButtonDisabled = true
	})

	ctx.NewActionWithValue("shade", key)
}

// onDismissToast is a callback function to call the dismiss action to hide any toast present in the UI.
func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

// onKeyDown is a callback function that enables the UI controlling using the keyboards keys.
func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	// If the key is Escape/Esc, cast new dismiss action.
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}
}

// onScroll is a callback function, that is called on any scroll in the UI. New scroll action is then called just after.
func (c *Content) onScroll(ctx app.Context, e app.Event) {
	ctx.NewAction("scroll")
}

// onSearch is a callback function that helps to order, show and render the searched (matching) user(s) in the UI.
func (c *Content) onSearch(ctx app.Context, e app.Event) {
	// Fetch the requested string to compare with the existing users map.
	val := ctx.JSSrc().Get("value").String()

	// Strings longer than 12 characters are ignored.
	if len(val) > 12 || len(val) < 2 {
		return
	}

	// Cast new search action.
	ctx.NewActionWithValue("search", val)
}

func (c *Content) onMouseEnter(ctx app.Context, e app.Event) {
	//ctx.JSSrc().Get("classList").Call("add", "underline")
	ctx.JSSrc().Get("style").Call("setProperty", "font-size", "1.85rem")
}

func (c *Content) onMouseLeave(ctx app.Context, e app.Event) {
	//ctx.JSSrc().Get("classList").Call("remove", "underline")
	ctx.JSSrc().Get("style").Call("setProperty", "font-size", "1.75rem")
}
