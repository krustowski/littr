// The stats (app's statistics) view and view-controllers logic package.
package stats

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

	flowStats map[string]int
	userStats map[string]models.UserStat

	nicknames []string

	users map[string]models.User

	//searchString string

	toast common.Toast

	loaderShow bool
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("search", c.handleSearch)

	c.loaderShow = true
}

func (c *Content) OnNav(ctx app.Context) {
	if app.IsServer {
		return
	}

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		input := &common.CallInput{
			Method:      "GET",
			Url:         "/api/v1/stats",
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		type dataModel struct {
			FlowStats map[string]int             `json:"flow_stats"`
			UserStats map[string]models.UserStat `json:"user_stats"`
			Users     map[string]models.User     `json:"users"`
		}

		output := &common.Response{Data: &dataModel{}}

		// fetch the stats
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code == 401 {
			toast.Text(common.ERR_LOGIN_AGAIN).Link("/logout").Type(common.TTYPE_INFO).Dispatch()

			if err := ctx.LocalStorage().Set("user", ""); err != nil {
				toast.Text(common.ErrLocalStorageUserLoad).Link("/logout").Type(common.TTYPE_INFO).Dispatch()
			}

			if err := ctx.LocalStorage().Set("authGranted", false); err != nil {
				toast.Text(common.ErrLocalStorageUserLoad).Link("/logout").Type(common.TTYPE_INFO).Dispatch()
			}
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

		ctx.Dispatch(func(ctx app.Context) {
			c.users = data.Users
			c.flowStats = data.FlowStats
			c.userStats = data.UserStats

			c.loaderShow = false
		})
	})
}
