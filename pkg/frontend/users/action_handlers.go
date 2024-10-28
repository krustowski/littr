package users

import (
	//"log"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// handleDismiss is an action handler function that ensures that any object to be dismissed is dismissed.
func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	// Dismiss the toast, enable buttons again, and hide the user modal.
	ctx.Dispatch(func(ctx app.Context) {
		c.toast.TText = ""
		c.usersButtonDisabled = false
		c.showUserPreviewModal = false
	})
}

// handleScroll is an action handler function that takes care of the action upon a generic scroll. More specially, it requests new items page, when the specified point/trigger is hit.
func (c *Content) handleScroll(ctx app.Context, a app.Action) {
	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// Get the anchoring element on the bottom of the page.
		elem := app.Window().GetElementByID("page-end-anchor")

		// Get the square boundaries of such object.
		boundary := elem.JSValue().Call("getBoundingClientRect")

		// Convert the bottom boundary to integer.
		bottom := boundary.Get("bottom").Int()

		// Get the height of the current display.
		_, height := app.Window().Size()

		// If the bottom-height difference is less than zero, the pagination end has not been hit yet, and the scroll processing has not been enabled so far: continue with the new-page fetch procedure.
		if bottom-height < 0 && !c.paginationEnd && !c.processingScroll {
			// Dispatch that a new scroll processing has just started.
			ctx.Dispatch(func(ctx app.Context) {
				c.processingScroll = true
			})

			// Get the page number.
			pageNo := c.pageNo

			// Compose the API call payload to fetch more pages.
			input := &common.CallInput{
				Method: "GET",
				Url:    "/api/v1/users",
				Data:   nil,
				PageNo: pageNo,
			}

			// Declare the response data model.
			type dataModel struct {
				Users     map[string]models.User     `json:"users"`
				Code      int                        `json:"code"`
				User      models.User                `json:"user"`
				UserStats map[string]models.UserStat `json:"user_stats"`
			}

			// Assign the data model to the API output object.
			output := &common.Response{Data: &dataModel{}}

			// Call the API to fetch more data pages.
			if ok := common.FetchData(input, output); !ok {
				toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			// Check if the HTTP 401 is present, if so, link the user to the logout route and terminate the procedure.
			if output.Code == 401 {
				toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Dispatch(c, dispatch)
				return
			}

			// Assert the type pointer to the data model.
			data, ok := output.Data.(*dataModel)
			if !ok {
				toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
				return
			}

			// Debugging outputs.
			//log.Printf("c.users: %d\n", len(c.users))
			//log.Printf("response.Users: %d\n", len(data.Users))

			// Loop over users and toggle them all to be searched for manually.
			for k, u := range data.Users {
				u.Searched = true
				data.Users[k] = u
			}

			// Get the current users map.
			users := c.users

			// Patch the users nil map.
			if users == nil {
				users = make(map[string]models.User)
			}

			// Loop the just-fetched users again, now to find out who is a new one in the list.
			for key, user := range data.Users {
				// Such user is already added in the users map. Get another one.
				if _, found := users[key]; found {
					continue
				}

				// Add new user to the map.
				users[key] = user
			}

			// Dispatch the changes to reflect the reality in the UI.
			ctx.Dispatch(func(ctx app.Context) {
				c.pageNo++
				c.users = users
				c.userStats = data.UserStats
				c.processingScroll = false
			})
			return
		}
		return
	})
}

// handleSearch is an action handler function that takes care of the user search tool.
func (c *Content) handleSearch(ctx app.Context, a app.Action) {
	// Fetch and assert the action value (the search string).
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		// Get the current users map.
		users := c.users

		// Loop over the users map, mark everyone as not-searched, and calculate the possible mathings.
		for key, user := range users {
			// Mark the user as not-searched-for.
			user.Searched = false

			// Use the lowercase to enable more flexible searching experience.
			lval := strings.ToLower(val)
			lkey := strings.ToLower(key)

			// Compare the lowercased value to the lowercased key.
			if strings.Contains(lkey, lval) {
				//log.Println(key)
				// Mark the user as searched if the match was found.
				user.Searched = true
			}

			// Update the user in the users map.
			users[key] = user
		}

		// Dispatch the users map changes to match the reality for the UI.
		ctx.Dispatch(func(ctx app.Context) {
			c.users = users
			c.loaderShow = false
		})
		return
	})
}

// handleToggle is an action handler that takes care of user follow toggling.
func (c *Content) handleToggle(ctx app.Context, a app.Action) {
	// Fetch the requested ID (nickname) and assert it to string.
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	// Fetch the current user from the Content struct. Get their flowList.
	user := c.user
	flowList := user.FlowList

	// If the requested user is already shaded, we have no job there.
	if c.user.ShadeList[key] {
		return
	}

	// Patch the nil flowList map.
	// Assign the following of oneself explicitly for the core app functions to work properly.
	if flowList == nil {
		flowList = make(map[string]bool)
		flowList[user.Nickname] = true
	}

	// Look for the key (counterpart's nickname) in the current flowList. Unfollow them if found. Follow the otherwise.
	// Assign the following explicitly by default (because we cannot untoggle the follow first, when the counterpart's record is not in the map yet).
	if value, found := flowList[key]; found {
		flowList[key] = !value
	} else {
		flowList[key] = true
	}

	// Also, ensure that the system account is followed by default too. Always.
	flowList["system"] = true

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// do not save new flow user to local var until it is saved on backend
		//flowRecords := append(c.flowRecords, flowName)

		// Prepare the request body data structure.
		payload := struct {
			FlowList map[string]bool `json:"flow_list"`
		}{
			FlowList: flowList,
		}

		// Compose the API call input payload.
		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + user.Nickname + "/lists",
			Data:        payload,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		// Prepare the blank API response output object.
		output := &common.Response{}

		// Patch the current user's flowList.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Check for the HTTP 200/201 response code(s), otherwise print the API response message in the toast.
		if output.Code != 200 && output.Code != 201 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Now, we can update the current user's flowList on the frontend too.
		// Update the flowList and update the user struct in the LocalStorage.
		user.FlowList = flowList
		ctx.LocalStorage().Set("user", user)

		// Dispatch the changes to match the reality for the UI to rerender.
		ctx.Dispatch(func(ctx app.Context) {
			c.usersButtonDisabled = false

			// Update the current user.
			c.users[user.Nickname] = user
			c.user = user
			c.user.FlowList = flowList
		})
		return
	})
}

// handleUserPreview is an action handler function that enables the showing of the user modal.
func (c *Content) handleUserPreview(ctx app.Context, a app.Action) {
	// Fetch the requested ID (nickname) and assert it type string.
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		// Fetch such user requested.
		user := c.users[val]

		// Dispatch the changes for the UI = enable showing of the requested user's modal.
		ctx.Dispatch(func(ctx app.Context) {
			c.showUserPreviewModal = true
			c.userInModal = user
		})
	})
	return
}
