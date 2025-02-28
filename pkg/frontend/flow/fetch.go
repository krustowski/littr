package flow

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type pageOptions struct {
	PageNo   int `default:0`
	Context  app.Context
	CallerID string

	SinglePost bool `default:"false"`
	UserFlow   bool `default:"false"`

	SinglePostID string `default:""`
	UserFlowNick string `default:""`

	Hashtag string `default:""`

	HideReplies bool `default:"false"`
}

func (c *Content) fetchFlowPage(opts pageOptions) (*map[string]models.Post, *map[string]models.User) {
	ctx := opts.Context
	pageNo := opts.PageNo

	toast := common.Toast{AppContext: &ctx}

	/*if opts.Context == (app.Context{}) {
		toast.Text("app context pointer cannot be nil").Type("error").Dispatch(c, dispatch)
		return nil, nil
	}*/

	//pageNo := c.pageNoToFetch
	if c.refreshClicked {
		pageNo = 0
	}
	//pageNoString := strconv.FormatInt(int64(pageNo), 10)

	url := "/api/v1/posts"
	if opts.UserFlow || opts.SinglePost || opts.Hashtag != "" {
		if opts.SinglePostID != "" {
			url += "/" + opts.SinglePostID
		}

		if opts.UserFlowNick != "" {
			//url += "/user/" + opts.UserFlowNick
			url = "/api/v1/users/" + opts.UserFlowNick + "/posts"
		}

		if opts.Hashtag != "" {
			url = "/api/v1/posts/hashtags/" + opts.Hashtag
		}

		if opts.SinglePostID == "" && opts.UserFlowNick == "" && opts.Hashtag == "" {
			toast.Text(common.ERR_INVALID_REQ_PARAMS).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				c.refreshClicked = false
			})
			return nil, nil
		}
	}

	input := &common.CallInput{
		Method:      "GET",
		Url:         url,
		Data:        nil,
		CallerID:    c.user.Nickname,
		PageNo:      pageNo,
		HideReplies: c.hideReplies,
	}

	type dataModel struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Code  int                    `json:"code"`
		Key   string                 `json:"key"`
	}

	output := &common.Response{Data: &dataModel{}}

	if ok := common.FetchData(input, output); !ok {
		toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			c.refreshClicked = false
		})
		return nil, nil
	}

	if output.Code == 401 {
		ctx.LocalStorage().Set("user", "")
		ctx.LocalStorage().Set("authGranted", false)

		toast.Text(common.ERR_LOGIN_AGAIN).Type(common.TTYPE_INFO).Link("/logout").Keep().Dispatch()
		return nil, nil
	}

	if output.Code != 200 {
		toast.Text(output.Message).Type("error").Dispatch()
		return nil, nil
	}

	data, ok := output.Data.(*dataModel)
	if !ok {
		toast.Text("cannot get data").Type("error").Dispatch()
		return nil, nil
	}

	if len(data.Posts) < 1 && opts.UserFlowNick == "" {
		toast.Text(common.MSG_EMPTY_FLOW).Type(common.TTYPE_INFO).Link("/post").Dispatch()
	}

	if len(data.Posts) < 1 && opts.UserFlowNick != "" && c.user.FlowList[opts.UserFlowNick] {
		toast.Text(common.MSG_USER_HAS_NOT_POSTED).Type(common.TTYPE_INFO).Link("/users").Dispatch()
		//return nil, nil
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.refreshClicked = false

		c.key = data.Key
		if data.Key != "" {
			c.user = c.users[data.Key]
		}
	})

	return &data.Posts, &data.Users
}
