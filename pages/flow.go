package pages

import (
	"sort"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type FlowPage struct {
	app.Compo
}

type flowContent struct {
	app.Compo
	posts []Post
}

type Post struct {
	Author    string
	Content   string
	Timestamp int
}

func (p *FlowPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&flowContent{},
		&footer{},
	)
}

func (p *FlowPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("flow / littr")
}

func (c *flowContent) Render() app.UI {
	c.posts = []Post{
		{Author: "system", Content: "welcome onboard bruh, lit ngl", Timestamp: 1669997122},
		{Author: "krusty", Content: "idk sth ig", Timestamp: 1669997543},
	}

	sort.SliceStable(c.posts, func(i, j int) bool {
		return c.posts[i].Timestamp > c.posts[j].Timestamp
	})

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr flow"),
		app.P().Text("exclusive content coming frfr"),
		app.Div().Class("space"),

		app.Table().Class("border left-align").Body(
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("align-left").Text("author, content, timestamp"),
				),
			),
			app.TBody().Body(
				app.Range(c.posts).Slice(func(i int) app.UI {
					post := c.posts[i]

					return app.Tr().Body(
						app.Td().Class("align-left").Body(
							app.B().Text(post.Author).Class("deep-orange-text"),
							app.Div().Class("space"),
							app.Text(post.Content),
							app.Div().Class("space"),
							app.Text(post.Timestamp),
						),
					)
				}),
			),
		),
	)
}
