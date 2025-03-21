package flow

import (
	"strconv"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type URIParts struct {
	SinglePost   bool
	SinglePostID string
	UserFlow     bool
	UserFlowNick string
	Hashtag      string
}

func (c *Content) parseFlowURI(ctx app.Context) URIParts {
	parts := URIParts{
		SinglePost:   false,
		SinglePostID: "",
		UserFlow:     false,
		UserFlowNick: "",
		Hashtag:      "",
	}

	// Split the URI by '/'.
	url := strings.Split(ctx.Page().URL().Path, "/")

	// Into at least 4 fields. ( '' / 'flow' / 'posts' / '{ID}' )
	if len(url) > 3 && url[3] != "" {
		switch url[2] {
		case "posts":
			parts.SinglePost = true
			parts.SinglePostID = url[3]

		case "users":
			parts.UserFlow = true
			parts.UserFlowNick = url[3]

		case "hashtags":
			parts.Hashtag = url[3]
		}
	}

	isPost := true
	if _, err := strconv.Atoi(parts.SinglePostID); parts.SinglePostID != "" && err != nil {
		// prolly not a post ID, but an user's nickname
		isPost = false
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.isPost = isPost
		c.userFlowNick = parts.UserFlowNick
		c.singlePostID = parts.SinglePostID
		c.hashtag = parts.Hashtag
	})

	return parts
}
