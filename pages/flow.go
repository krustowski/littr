package pages

import (
	"log"

	"litter-go/backend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type FlowPage struct {
	app.Compo
}

type flowContent struct {
	app.Compo
	posts []backend.Post
}

func (p *FlowPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("flow / littr")
}

func (p *FlowPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&flowContent{},
		app.Div().Class("large-space"),
		app.Div().Class("large-space"),
		&footer{},
	)
}

func (c *flowContent) OnNav(ctx app.Context) {
	var posts []backend.Post

	ctx.Async(func() {
		if pp := backend.GetPosts(); pp != nil {
			posts = *pp
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.posts = posts
			log.Println("dispatch ends")
		})
	})
}

func (c *flowContent) Render() app.UI {

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr flow"),
		app.P().Text("exclusive content coming frfr"),
		app.Div().Class("space"),

		app.Table().Class("border left-align").Body(
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("align-left").Text("nickname, content, timestamp"),
				),
			),
			app.TBody().Body(
				app.Range(c.posts).Slice(func(i int) app.UI {
					post := c.posts[i]

					return app.Tr().Body(
						app.Td().Class("align-left").Body(
							app.B().Text(post.Nickname).Class("deep-orange-text"),
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
