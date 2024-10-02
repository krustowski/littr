package polls

import (
	"encoding/json"
	"log"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	eventListener func()

	loggedUser string
	user       models.User

	loaderShow bool

	toastShow bool
	toastText string
	toastType string

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
	// show loader
	c.loaderShow = true
	var toastText string
	var toastType string

	ctx.Async(func() {
		pollsRaw := struct {
			Polls map[string]models.Poll `json:"polls"`
			Code  int                    `json:"code"`
			Users map[string]models.User `json:"users"`
			Key   string                 `json:"key"`
		}{}

		input := common.CallInput{
			Method:      "GET",
			Url:         "/api/v1/polls",
			Data:        nil,
			CallerID:    "",
			PageNo:      0,
			HideReplies: false,
		}

		if byteData, _ := common.CallAPI(input); byteData != nil {
			err := json.Unmarshal(*byteData, &pollsRaw)
			if err != nil {
				log.Println(err.Error())

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = err.Error()
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			toastText = "cannot fetch polls list"
			log.Println(toastText)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if pollsRaw.Code == 401 {
			toastText = "please log-in again"

			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		if len(pollsRaw.Polls) < 1 {
			toastText = "there is nothing here yet, be the first to create a poll!"
			toastType = "info"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastType = toastType
				c.toastShow = (toastText != "")
				c.loaderShow = false
			})
			return
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.user = pollsRaw.Users[pollsRaw.Key]
			c.loggedUser = c.user.Nickname

			c.pagination = 10
			c.pageNo = 1

			c.polls = pollsRaw.Polls

			c.pollsButtonDisabled = false
			c.loaderShow = false
		})
	})
	return
}

func (c *Content) OnMount(ctx app.Context) {
	ctx.Handle("vote", c.handleVote)
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("dismiss", c.handleDismiss)

	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1

	c.eventListener = app.Window().AddEventListener("scroll", c.onScroll)
	c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
}
