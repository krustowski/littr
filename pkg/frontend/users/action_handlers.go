package users

import (
	//"fmt"
	//"log"

	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
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

func (c *Content) handleLink(ctx app.Context, a app.Action) {
	common.HandleLink(ctx, a, "/flow/users/", "/flow/users/")
}

func (c *Content) handleMouseEnter(ctx app.Context, a app.Action) {
	common.HandleMouseEnter(ctx, a)
}

func (c *Content) handleMouseLeave(ctx app.Context, a app.Action) {
	common.HandleMouseLeave(ctx, a)
}

func (c *Content) handlePrivateMode(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.usersButtonDisabled = true
		c.userButtonDisabled = true
	})

	callback := func(updateUser bool) {
		c.usersButtonDisabled = false
		c.userButtonDisabled = false

		/*if updateUser {
			updatee := c.users[key]
			updatee.RequestList[c.user.Nickname] = !updatee.RequestList[c.user.Nickname]
			c.users[key] = updatee
		}*/
	}

	common.HandlePrivateMode(ctx, a, c.users[key], callback)
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
				toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
				return
			}

			// Check if the HTTP 401 is present, if so, link the user to the logout route and terminate the procedure.
			if output.Code == 401 {
				toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Dispatch()
				return
			}

			// Assert the type pointer to the data model.
			data, ok := output.Data.(*dataModel)
			if !ok {
				toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch()
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
		}
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
	})
}

// handleToggle is an action handler that takes care of user follow toggling.
func (c *Content) handleToggle(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.userButtonDisabled = true
		c.usersButtonDisabled = true
	})

	callback := func(updateUser bool) {
		c.userButtonDisabled = false
		c.usersButtonDisabled = false

		c.user.Searched = true

		ctx.GetState(common.StateNameUser, &c.user)
	}

	common.HandleToggleFollow(ctx, a, callback)
}

// handleUserPreview is an action handler function that enables the showing of the user modal.
func (c *Content) handleUserPreview(ctx app.Context, a app.Action) {
	// Fetch the requested ID (nickname) and assert it type string.
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	// Dispatch the changes for the UI = enable showing of the requested user's modal.
	ctx.Dispatch(func(ctx app.Context) {
		c.showUserPreviewModal = true
		c.userInModal = c.users[val]
	})
}

// handleUserShade is an action handler function that enables one to shade other accounts.
func (c *Content) handleUserShade(ctx app.Context, a app.Action) {
	// Fetch the requested ID (nickname) and assert it type string.
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.userButtonDisabled = true
		c.usersButtonDisabled = true
	})

	callback := func(updateUser bool) {
		c.userButtonDisabled = false
		c.usersButtonDisabled = false

		if updateUser {
			updatedUser := c.users[key]

			updatedUser.FlowList[c.user.Nickname] = false
			c.user.ShadeList[key] = true

			c.users[key] = updatedUser
		}

		ctx.GetState(common.StateNameUser, &c.user)
	}

	common.HandleUserShade(ctx, a, c.users[key], callback)
}
