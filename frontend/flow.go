package frontend

import (
	"encoding/json"
	"log"

	"litter-go/config"
	"litter-go/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type FlowPage struct {
	app.Compo
}

type flowContent struct {
	app.Compo

	loaderShow bool
	loggedUser string

	posts map[string]models.Post
}

func (p *FlowPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("flow / littr")
}

func (p *FlowPage) Render() app.UI {
	return app.Div().Body(
		app.Body().Class("dark"),
		&header{},
		&footer{},
		&flowContent{},
	)
}

func (c *flowContent) OnMount(ctx app.Context) {
	ctx.LocalStorage().Get("userName", &c.loggedUser)
}

func (c *flowContent) OnNav(ctx app.Context) {
	c.loaderShow = true

	ctx.Async(func() {
		postsRaw := struct {
			Posts map[string]models.Post `json:"posts"`
		}{}

		if byteData, _ := litterAPI("GET", "/api/flow", nil); byteData != nil {
			err := json.Unmarshal(*byteData, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				return
			}
		} else {
			log.Println("cannot fetch post flow list")
			return
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.posts = postsRaw.Posts

			c.loaderShow = false
			log.Println("dispatch ends")
		})
	})
}

func (c *flowContent) Render() app.UI {
	loaderActiveClass := ""
	if c.loaderShow {
		loaderActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr flow").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("exclusive content incoming frfr"),
		app.Div().Class("space"),

		app.Table().Class("border left-align").Body(
			// table header
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("align-left").Text("nickname, content, timestamp"),
				),
			),

			// table body
			app.TBody().Body(
				app.Range(c.posts).Map(func(key string) app.UI {
					post := c.posts[key]

					return app.Tr().Body(
						app.Td().Class("align-left").Body(
							app.P().Body(
								app.B().Text(post.Nickname).Class("deep-orange-text"),
							),

							app.P().Body(
								app.Text(post.Content),
							),

							app.Div().Class("row").Body(
								app.Div().Class("max").Body(
									app.Text(post.Timestamp.Format("Jan 02, 2006; 15:04:05")),
								),
								app.If(c.loggedUser == post.Nickname,
									app.B().Text(post.ReactionCount),
									app.Button().Class("transparent circle").Body(
										app.I().Text("delete"),
									),
								).Else(
									app.B().Text(post.ReactionCount),
									app.Button().Class("transparent circle").Body(
										app.I().Text("star"),
									),
								),
							),
						),
					)
				}),
			),
		),

		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.A().Class("loader center large deep-orange"+loaderActiveClass),
		),
	)
}
