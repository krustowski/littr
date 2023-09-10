package frontend

import (
	"encoding/json"
	"log"
	"sort"

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

	toastShow bool
	toastText string

	interactedPostKey string

	posts map[int]models.Post
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

func (c *flowContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
}

func (c *flowContent) onClickDelete(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		key := ctx.JSSrc().Get("id").String()
		log.Println(key)

		if key == "" {
			return
		}

		var author string
		ctx.LocalStorage().Get("userName", &author)

		interactedPost := c.posts[key]

		if _, ok := litterAPI("DELETE", "/api/flow", interactedPost); !ok {
			c.toastShow = true
			c.toastText = "backend error: cannot delete a post"
			log.Println("cannot delete a post via API!")
			return
		}

		c.toastShow = false
		ctx.Navigate("/flow")
	})
}

func (c *flowContent) onClickStar(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		key := ctx.JSSrc().Get("id").String()
		log.Println(key)

		if key == "" {
			return
		}

		var author string
		ctx.LocalStorage().Get("userName", &author)

		interactedPost := c.posts[key]
		interactedPost.ReactionCount++

		// add new post to backend struct
		if _, ok := litterAPI("PUT", "/api/flow", interactedPost); !ok {
			c.toastShow = true
			c.toastText = "backend error: cannot rate a post"
			log.Println("cannot rate a post via API!")
			return
		}

		c.toastShow = false
		ctx.Navigate("/flow")
	})
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

		// order posts by timestamp DESC
		sort.SliceStable(postsRaw.Posts, func(i, j int) bool {
			return postsRaw.Posts[i].Timestamp.After(postsRaw.Posts[j].Timestamp)
		})

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

	toastActiveClass := ""
	if c.toastShow {
		toastActiveClass = " active"
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr flow").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("exclusive content incoming frfr"),
		app.Div().Class("space"),

		app.A().OnClick(c.dismissToast).Body(
			app.Div().Class("toast red10 white-text top"+toastActiveClass).Body(
				app.I().Text("error"),
				app.Span().Text(c.toastText),
			),
		),

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
									app.Text(post.Timestamp.Format("Jan 02, 2006; 15:04:05 -0700")),
								),
								app.If(c.loggedUser == post.Nickname,
									app.B().Text(post.ReactionCount),
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickDelete).Body(
										app.I().Text("delete"),
									),
								).Else(
									app.B().Text(post.ReactionCount),
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickStar).Body(
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
			app.Div().Class("loader center large deep-orange"+loaderActiveClass),
		),
	)
}
