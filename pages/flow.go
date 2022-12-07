package pages

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"litter-go/backend"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type FlowPage struct {
	app.Compo
}

type flowContent struct {
	app.Compo

	loaderShow bool

	posts []backend.Post
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

func flowAPI() *[]byte {
	// push requests use PUT method
	req, err := http.NewRequest("GET", "/api/flow", nil)
	if err != nil {
		log.Print(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		log.Print(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}

	return &data
}

func (c *flowContent) OnNav(ctx app.Context) {
	ctx.Async(func() {
		var postsRaw backend.Posts

		if uu := flowAPI(); uu != nil {
			err := json.Unmarshal(*uu, &postsRaw)
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
							app.Text(post.Timestamp.Format("Jan 02, 2006; 15:04:05")),
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
