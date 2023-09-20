package frontend

import (
	"encoding/json"
	"log"
	"sort"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type FlowPage struct {
	app.Compo
}

type flowContent struct {
	app.Compo

	loaderShow bool
	loaderShowImage bool

	loggedUser string
	user       models.User

	toastShow bool
	toastText string

	interactedPostKey string

	postKey     string
	posts       map[string]models.Post
	sortedPosts []models.Post
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

func (c *flowContent) onLoadStartImage(ctx app.Context, e app.Event) {
	log.Println("media started loading...")
	c.loaderShowImage = true
}
func (c *flowContent) onLoadedDataImage(ctx app.Context, e app.Event) {
	log.Println("media loaded")
	c.loaderShowImage = false
}

func (c *flowContent) onClickDelete(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("delete", key)
}

func (c *flowContent) handleDelete(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	c.postKey = key

	ctx.Async(func() {
		var toastText string = ""

		key := c.postKey
		interactedPost := c.posts[key]

		if _, ok := litterAPI("DELETE", "/api/flow", interactedPost); !ok {
			toastText = "backend error: cannot delete a post"
		}

		ctx.Dispatch(func(ctx app.Context) {
			delete(c.posts, key)

			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
	})
}

func (c *flowContent) onClickStar(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	ctx.NewActionWithValue("star", key)
}

func (c *flowContent) handleStar(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	// runs on the main UI goroutine via a component ActionHandler
	post := c.posts[key]
	post.ReactionCount++
	c.posts[key] = post
	c.postKey = key

	ctx.Async(func() {
		//var author string
		var toastText string = ""

		//key := ctx.JSSrc().Get("id").String()
		key := c.postKey
		//author = c.user.Nickname

		interactedPost := c.posts[key]
		//interactedPost.ReactionCount++

		// add new post to backend struct
		if _, ok := litterAPI("PUT", "/api/flow", interactedPost); !ok {
			toastText = "backend error: cannot rate a post"
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.posts[key] = interactedPost
			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
	})
}

func (c *flowContent) OnMount(ctx app.Context) {
	var enUser string
	var user models.User
	var toastText string = ""

	ctx.Handle("star", c.handleStar)
	ctx.Handle("delete", c.handleDelete)

	ctx.LocalStorage().Get("user", &enUser)
	// decode, decrypt and unmarshal the local storage string
	if err := prepare(enUser, &user); err != nil {
		toastText = "frontend decoding/decryption failed: " + err.Error()
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.user = user
		c.loggedUser = user.Nickname
		c.toastText = toastText
		c.toastShow = (toastText != "")
	})
	return
}

func (c *flowContent) OnNav(ctx app.Context) {
	// show loader
	c.loaderShow = true
	c.loaderShowImage = true

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
			//c.sortedPosts = posts

			c.loaderShow = false
			c.loaderShowImage = false
			log.Println("dispatch ends")
		})
		return
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

	var sortedPosts []models.Post

	for _, sortedPost := range c.posts {
		sortedPosts = append(sortedPosts, sortedPost)
	}

	// order posts by timestamp DESC
	sort.SliceStable(sortedPosts, func(i, j int) bool {
		return sortedPosts[i].Timestamp.After(sortedPosts[j].Timestamp)
	})

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
				//app.Range(c.posts).Map(func(key string) app.UI {
				app.Range(sortedPosts).Slice(func(idx int) app.UI {
					//post := c.sortedPosts[idx]
					post := sortedPosts[idx]
					key := post.ID

					// only show posts of users in one's flowList
					if !c.user.FlowList[post.Nickname] && post.Nickname != "system" {
						return nil
					}

					return app.Tr().Body(
						app.Td().Class("align-left").Body(
							app.P().Body(
								app.B().Text(post.Nickname).Class("deep-orange-text"),
							),

							app.If(post.Type == "fig",
								app.Article().Class("medium no-padding transparent").OnScroll(c.onLoadStartImage).Body(
									app.If(c.loaderShowImage,
										app.Div().Class("small-space"),
											app.Div().Class("loader center large deep-orange active"),
									),
									app.Img().Class("lazy no-padding priamry absolute center middle").Src(post.Content).Style("max-width", "100%").Style("max-height", "100%").OnLoadStart(c.onLoadStartImage).OnLoadedData(c.onLoadedDataImage).Attr("loading", "lazy").On("onloadstart", c.onLoadStartImage).OnScroll(c.onLoadStartImage),
									
								),
							).Else(
								app.P().Body(
									app.Text(post.Content),
								),
							),

							app.Div().Class("row").Body(
								app.Div().Class("max").Body(
									app.Text(post.Timestamp.Format("Jan 02, 2006; 15:04:05 -0700")),
								),
								app.If(c.user.Nickname == post.Nickname,
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
