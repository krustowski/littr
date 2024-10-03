package polls

import (
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

	toastShow bool
	toastText string
	toastType string
	toast     Toast

	paginationEnd bool
	pagination    int
	pageNo        int

	pollKey                    string
	interactedPollKey          string
	deleteModalButtonsDisabled bool
	deletePollModalShow        bool

	polls map[string]models.Poll

	pollsButtonDisabled bool

	keyDownEventListener func()
}

func (c *Content) OnNav(ctx app.Context) {
	toast := Toast{AppContext: &ctx}

	ctx.Async(func() {
		input := common.CallInput{
			Method:      "GET",
			Url:         "/api/v1/polls",
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		response := &struct {
			Polls map[string]models.Poll `json:"polls"`
			Code  int                    `json:"code"`
			User  models.User            `json:"user"`
		}{}

		// call the API to fetch the data
		if ok := common.CallAPI(input, response); !ok {
			toast.Text("cannot fetch polls list").Dispatch(c)
			return
		}

		if response.Code == 401 {
			// void user's session indicators in LocalStorage
			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			toast.Text("please log-in again").Dispatch(c)
			return
		}

		if len(response.Polls) < 1 {
			toast.Text("there is no poll yet, be the first to create one!").Type("info").Link("/post").Dispatch(c)

			ctx.Dispatch(func(ctx app.Context) {
				c.loaderShow = false
			})
			return
		}

		// storing the HTTP response in Content fields:
		ctx.Dispatch(func(ctx app.Context) {
			c.user = response.User
			//c.loggedUser = c.user.Nickname

			c.pagination = 10
			c.pageNo = 1

			c.polls = response.Polls

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
