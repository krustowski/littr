// The users view and view-controllers logic package.
package users

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	users map[string]models.User

	user        models.User
	userInModal models.User

	flowStats map[string]int
	userStats map[string]models.UserStat

	userButtonDisabled bool

	loaderShow bool

	paginationEnd bool
	pagination    int
	pageNo        int

	toast common.Toast

	usersButtonDisabled  bool
	showUserPreviewModal bool

	processingScroll bool
}

func (c *Content) OnNav(ctx app.Context) {
	if app.IsServer {
		return
	}

	ctx.GetState(common.StateNameUser, &c.user)

	// show loader
	c.loaderShow = true
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		input := &common.CallInput{
			Method: "GET",
			Url:    "/api/v1/users",
			Data:   nil,
			PageNo: 0,
		}

		type dataModel struct {
			User      models.User                `json:"user"`
			Users     map[string]models.User     `json:"users"`
			UserStats map[string]models.UserStat `json:"user_stats"`
			Code      int                        `json:"code"`
		}

		output := &common.Response{Data: &dataModel{}}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code == 401 {
			if err := ctx.LocalStorage().Set("user", ""); err != nil {
				return
			}
			if err := ctx.LocalStorage().Set("authGranted", false); err != nil {
				return
			}

			toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Toggle all users to be "searched for" on init manually
		for k, u := range data.Users {
			u.Searched = true
			data.Users[k] = u
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.user = data.User
			c.users = data.Users
			c.userStats = data.UserStats

			c.pagination = 25
			c.pageNo = 1

			c.loaderShow = false
		})
	})
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("dismiss", c.handleDismiss)
	ctx.Handle("preview", c.handleUserPreview)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("search", c.handleSearch)
	ctx.Handle("toggle", c.handleToggle)
	ctx.Handle("flow-link-click", c.handleLink)

	ctx.Handle("mouse-enter", c.handleMouseEnter)
	ctx.Handle("mouse-leave", c.handleMouseLeave)

	ctx.Handle("user", c.handleUserPreview)
	ctx.Handle("nickname-click", c.handleUserPreview)

	ctx.Handle("ask", c.handlePrivateMode)
	ctx.Handle("cancel", c.handlePrivateMode)
	ctx.Handle("follow", c.handleToggle)
	ctx.Handle("shade", c.handleUserShade)
	ctx.Handle("unfollow", c.handleToggle)

	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1
}
