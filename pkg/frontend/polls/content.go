// The polls view and view-controllers logic package.
package polls

import (
	"strings"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	scrollEventListener func()

	loggedUser string
	user       models.User

	loaderShow bool

	toast common.Toast

	paginationEnd bool
	pagination    int
	pageNo        int

	pollKey                    string
	interactedPollKey          string
	deleteModalButtonsDisabled bool
	deletePollModalShow        bool

	polls map[string]models.Poll

	pollsButtonDisabled bool

	processingScroll bool

	keyDownEventListener func()

	singlePollID string
}

func (c *Content) OnNav(ctx app.Context) {
	toast := common.Toast{AppContext: &ctx}

	url := strings.Split(ctx.Page().URL().Path, "/")

	var singlePollID string

	if len(url) > 3 && url[3] != "" {
		switch url[2] {
		case "poll":
			singlePollID = url[3]
		}
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.singlePollID = singlePollID
	})

	ctx.Async(func() {
		input := &common.CallInput{
			Method: "GET",
			Url: func() string {
				if singlePollID != "" {
					return "/api/v1/polls/" + singlePollID
				}

				return "/api/v1/polls"
			}(),
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		type dataModel struct {
			Poll  models.Poll            `json:"poll"`
			Polls map[string]models.Poll `json:"polls"`
			User  models.User            `json:"user"`
		}

		output := &common.Response{Data: &dataModel{}}

		// call the API to fetch the data
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code == 401 {
			// void user's session indicators in LocalStorage
			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Dispatch(c, dispatch)
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if data.Polls == nil {
			data.Polls = make(map[string]models.Poll)
			data.Polls[data.Poll.ID] = data.Poll
		}

		if len(data.Polls) < 1 {
			ctx.Dispatch(func(ctx app.Context) {
				c.loaderShow = false
			})

			toast.Text(common.MSG_NO_POLL_TO_SHOW).Type(common.TTYPE_INFO).Link("/post").Dispatch(c, dispatch)
			return
		}

		// storing the HTTP response in Content fields:
		ctx.Dispatch(func(ctx app.Context) {
			c.user = data.User
			//c.loggedUser = c.user.Nickname

			c.pagination = 25
			c.pageNo = 1

			c.polls = data.Polls

			c.pollsButtonDisabled = false
			c.loaderShow = false
		})
	})
	return
}

func (c *Content) OnMount(ctx app.Context) {
	// action handlers
	ctx.Handle("vote", c.handleVote)
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("dismiss", c.handleDismiss)

	// show loader
	c.loaderShow = true

	// pagination
	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1

	// tweaked EventListeners (may cause memory leaks when not closed properly!)
	c.scrollEventListener = app.Window().AddEventListener("scroll", c.onScroll)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}
