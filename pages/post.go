package pages

import (
	"litter-go/backend"
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type PostPage struct {
	app.Compo
}

type postContent struct {
	app.Compo
	newPost string
}

func (p *PostPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("post / littr")
}

func (p *PostPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&postContent{},
		&footer{},
	)
}

func (c *postContent) onClick(ctx app.Context, e app.Event) {
	if c.newPost != "" {
		log.Println(c.newPost)
		// add new post to backend struct
		backend.AddPost(c.newPost)
		ctx.Navigate("/flow")
	}
}

func (c *postContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.H5().Text("new flow post"),
		app.P().Text("pleasure to be lit"),

		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Name("newPost").OnChange(c.ValueTo(&c.newPost)),
			app.Label().Text("contents"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("post").OnClick(c.onClick),
	)
}
