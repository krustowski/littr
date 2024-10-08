package users

import (
	"log"
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toast.TText = ""
		c.usersButtonDisabled = false
		c.showUserPreviewModal = false
	})
}

func (c *Content) handleScroll(ctx app.Context, a app.Action) {
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		if bottom-height < 0 && !c.paginationEnd && !c.processingScroll {
			ctx.Dispatch(func(ctx app.Context) {
				c.processingScroll = true
			})

			pageNo := c.pageNo

			input := common.CallInput{
				Method: "GET",
				Url:    "/api/v1/users",
				Data:   nil,
				PageNo: pageNo,
			}

			response := &struct {
				Users     map[string]models.User     `json:"users"`
				Code      int                        `json:"code"`
				User      models.User                `json:"user"`
				UserStats map[string]models.UserStat `json:"user_stats"`
			}{}

			// call the API to fetch the data
			if ok := common.CallAPI(input, response); !ok {
				toast.Text("cannot fetch users list").Dispatch(c, dispatch)
				return
			}

			if response.Code == 401 {
				toast.Text("please log-in again").Link("/logout").Dispatch(c, dispatch)
				return
			}

			log.Printf("c.users: %d\n", len(c.users))
			log.Printf("response.Users: %d\n", len(response.Users))

			// manually toggle all users to be "searched for" on init
			for k, u := range response.Users {
				u.Searched = true
				response.Users[k] = u
			}

			users := c.users
			if users == nil {
				users = make(map[string]models.User)
			}

			for key, user := range response.Users {
				if _, found := users[key]; found {
					continue
				}

				users[key] = user
			}

			ctx.Dispatch(func(ctx app.Context) {
				c.pageNo++
				c.users = users
				c.userStats = response.UserStats
				c.processingScroll = false
			})
			return
		}
	})
}

func (c *Content) handleSearch(ctx app.Context, a app.Action) {
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		users := c.users

		// iterate over calculated stats' "rows" and find matchings
		for key, user := range users {
			//user := users[key]
			user.Searched = false

			// use lowecase to search across UNICODE letters
			lval := strings.ToLower(val)
			lkey := strings.ToLower(key)

			if strings.Contains(lkey, lval) {
				log.Println(key)
				user.Searched = true
			}

			users[key] = user
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.users = users

			c.loaderShow = false
		})
		return
	})
}

func (c *Content) handleToggle(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	user := c.user
	flowList := user.FlowList

	if c.user.ShadeList[key] {
		return
	}

	if flowList == nil {
		flowList = make(map[string]bool)
		flowList[user.Nickname] = true
		//c.user.FlowList = flowList
	}

	if value, found := flowList[key]; found {
		flowList[key] = !value
	} else {
		flowList[key] = true
	}

	flowList["system"] = true

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		// do not save new flow user to local var until it is saved on backend
		//flowRecords := append(c.flowRecords, flowName)

		payload := struct {
			FlowList map[string]bool `json:"flow_list"`
		}{
			FlowList: flowList,
		}

		input := common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + user.Nickname + "/lists",
			Data:        payload,
			CallerID:    user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		response := struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}{}

		if ok := common.CallAPI(input, &response); !ok {
			toast.Text("generic backend error").Type("error").Dispatch(c, dispatch)
			return
		}

		if response.Code != 200 && response.Code != 201 {
			toast.Text("user update failed: "+response.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		user.FlowList = flowList
		ctx.LocalStorage().Set("user", user)

		ctx.Dispatch(func(ctx app.Context) {
			c.usersButtonDisabled = false

			c.users[user.Nickname] = user
			c.user = user
			c.user.FlowList = flowList
		})
	})
}

func (c *Content) handleUserPreview(ctx app.Context, a app.Action) {
	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		user := c.users[val]

		ctx.Dispatch(func(ctx app.Context) {
			c.showUserPreviewModal = true
			c.userInModal = user
		})
	})
	return
}
