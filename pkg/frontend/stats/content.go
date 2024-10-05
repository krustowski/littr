package stats

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	flowStats map[string]int
	userStats map[string]models.UserStat

	nicknames []string

	users map[string]models.User

	searchString string

	toast common.Toast

	loaderShow bool
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("search", c.handleSearch)

	c.loaderShow = true
}

func (c *Content) OnNav(ctx app.Context) {
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		payload := struct {
			FlowStats map[string]int             `json:"flow_stats"`
			UserStats map[string]models.UserStat `json:"user_stats"`
			Users     map[string]models.User     `json:"users"`
			Code      int                        `json:"code"`
		}{}

		input := common.CallInput{
			Method:      "GET",
			Url:         "/api/v1/stats",
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		// fetch the stats
		if ok := common.CallAPI(input, &payload); !ok {
			toast.Text("cannot fetch stats").Type("error").Dispatch(c, dispatch)
			return
		}

		if payload.Code == 401 {
			toast.Text("please log-in again").Type("info").Dispatch(c, dispatch)

			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.users = payload.Users
			c.flowStats = payload.FlowStats
			c.userStats = payload.UserStats

			c.loaderShow = false
		})
		return
	})
}
