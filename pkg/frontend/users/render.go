package users

import (
	"sort"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/organisms"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) processUsers() []models.User {
	keys := []string{}

	// prepare the keys array
	for key := range c.users {
		keys = append(keys, key)
	}

	// sort them keys
	sort.Strings(keys)

	// prepare the sorted users array
	sortedUsers := func() []models.User {
		var sorted []models.User

		for _, key := range keys {
			if !c.users[key].Searched {
				continue
			}

			sorted = append(sorted, c.users[key])
		}

		return sorted
	}()

	// prepare posts according to the actual pagination and pageNo
	pagedUsers := []models.User{}
	//pagedUsers := sortedUsers

	end := len(sortedUsers)
	start := 0

	stop := func(c *Content) int {
		var pos int

		if c.pagination > 0 {
			// (c.pageNo - 1) * c.pagination + c.pagination
			pos = c.pageNo * c.pagination
		}

		if pos > end {
			// kill the scrollEventListener (observers scrolling)
			//c.scrollEventListener()
			c.paginationEnd = true

			return (end)
		}

		if pos < 0 {
			return 0
		}

		return pos
	}(c)

	if end > 0 && stop > 0 {
		pagedUsers = sortedUsers[start:stop]
	}

	return pagedUsers
}

func (c *Content) sumRequests() int {
	var numOfReqs int = 0

	requestList := c.user.RequestList
	for _, state := range requestList {
		if state {
			numOfReqs++

			// We don't need to loop further as the number is always going to be
			// greater than zero henceforth.
			break
		}
	}

	return numOfReqs
}

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.If(c.user.RequestList != nil && c.sumRequests() > 0, func() app.UI {
			return app.Div().Body(
				&atoms.PageHeading{
					Title: "requests",
				},

				&organisms.UserRequests{
					LoggedUser:              c.user,
					Users:                   c.users,
					OnClickAllowActionName:  "allow",
					OnClickCancelActionName: "cancel",
					OnClickUserActionName:   "user",
					OnMouseEnterActionName:  "mouse-enter",
					OnMouseLeaveActionName:  "mouse-leave",
					ButtonsDisabled:         c.userButtonDisabled,
				},

				app.Div().Class("large-space"),
			)
		}),

		&atoms.PageHeading{
			Title: "flowers",
		},

		&organisms.ModalUserInfo{
			User:                      c.userInModal,
			Users:                     c.users,
			ShowModal:                 c.showUserPreviewModal,
			OnClickDismissActionName:  "dismiss",
			OnClickUserFlowActionName: "flow-link-click",
		},

		&atoms.SearchBar{
			ID:                 "user-search",
			OnSearchActionName: "search",
		},

		&organisms.UserFeed{
			LoggedUser:  c.user,
			SortedUsers: c.processUsers(),
			Users:       c.users,
			FlowStats:   c.flowStats,
			UserStats:   c.userStats,
			Pagination:  c.pagination,
			PageNo:      c.pageNo,
			//
			ButtonsDisabled: c.userButtonDisabled,
			LoaderShowImage: c.loaderShow,
			//
			OnClickUserActionName:      "user",
			OnClickUnfollowActionName:  "unfollow",
			OnClickAskActionName:       "ask",
			OnClickCancelActionName:    "cancel",
			OnClickFollowActionName:    "follow",
			OnClickPostCountActionName: "flow-link-click",
			OnClickNicknameActionName:  "nickname-click",
			OnClickShadeActionName:     "shade",
			OnMouseEnterActionName:     "mouse-enter",
			OnMouseLeaveActionName:     "mouse-leave",
		},

		&atoms.Loader{
			ID:         "page-end-anchor",
			ShowLoader: c.loaderShow,
		},
	)
}
