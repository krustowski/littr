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
	// Intantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	// Split the URI by '/'.
	url := strings.Split(ctx.Page().URL().Path, "/")

	var singlePollID string

	// At least three parts.
	// --> ''/'polls'/'ID'
	if len(url) > 2 {
		singlePollID = url[2]
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.singlePollID = singlePollID
	})

	ctx.Async(func() {
		// Compose the API input payload.
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

		// Declare the data model for the API response.
		type dataModel struct {
			Poll  models.Poll            `json:"poll"`
			Polls map[string]models.Poll `json:"polls"`
			User  models.User            `json:"user"`
		}

		// Prepare the API output object with assigned data model's pointer.
		output := &common.Response{Data: &dataModel{}}

		// Call the API to fetch the data.
		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		if output.Code == 401 {
			// Void user's session indicators in the LocalStorage.
			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Dispatch(c, dispatch)
			return
		}

		// Check if the response code is HTTP 200, otherwise print the API response message.
		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Assert the output data to the data model declared before.
		data, ok := output.Data.(*dataModel)
		if !ok {
			toast.Text(common.ERR_CANNOT_GET_DATA).Type(common.TTYPE_ERR).Dispatch(c, dispatch)
			return
		}

		// Check the contents of polls.
		if data.Polls == nil {
			data.Polls = make(map[string]models.Poll)
			data.Polls[data.Poll.ID] = data.Poll
		}

		// Less then one? Nothing to show then.
		if len(data.Polls) < 1 {
			ctx.Dispatch(func(ctx app.Context) {
				c.loaderShow = false
			})

			toast.Text(common.MSG_NO_POLL_TO_SHOW).Type(common.TTYPE_INFO).Link("/post").Dispatch(c, dispatch)
			return
		}

		// Store the HTTP response in the Content fields.
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
	// Action handlers.
	ctx.Handle("vote", c.handleVote)
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("dismiss", c.handleDismiss)

	// The loader.
	c.loaderShow = true

	// The pagination settings.
	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1

	// Tweaked EventListeners (may cause memory leaks when not closed properly!)
	c.scrollEventListener = app.Window().AddEventListener("scroll", c.onScroll)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}
