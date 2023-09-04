package frontend

import (
	"log"
	"time"

	"litter-go/config"
	"litter-go/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type PostPage struct {
	app.Compo
}

type postContent struct {
	app.Compo
	newPost string

	toastShow bool
	toastText string
}

func (p *PostPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("post / littr")
}

func (p *PostPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&postContent{},
	)
}

func (c *postContent) onClick(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		if c.newPost == "" {
			c.toastShow = true
			c.toastText = "post textarea must be filled"
			return
		}

		var author string
		ctx.LocalStorage().Get("userName", &author)

		// add new post to backend struct
		if _, ok := litterAPI("POST", "/api/flow", models.Post{
			Nickname:  author,
			Content:   c.newPost,
			Timestamp: time.Now(),
		}); !ok {
			c.toastShow = true
			c.toastText = "backend error: cannot add new post"
			log.Println("cannot post new post to API!")
			return
		}

		c.toastShow = false
		ctx.Navigate("/flow")
	})
}

func (c *postContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *postContent) Render() app.UI {
	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("add flow post").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("pleasure to be lit"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

		app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
			app.Textarea().Name("newPost").OnChange(c.ValueTo(&c.newPost)).AutoFocus(true),
			app.Label().Text("contents"),
		),
		app.Button().Class("responsive deep-orange7 white-text bold").Text("post").OnClick(c.onClick),
	)
}
