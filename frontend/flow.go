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

	eventListener func()

	loaderShow      bool
	loaderShowImage bool

	loggedUser string
	user       models.User

	toastShow bool
	toastText string

	interactedPostKey string

	paginationEnd bool
	pagination    int
	pageNo        int

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

func (c *flowContent) onScroll(ctx app.Context, e app.Event) {
       ctx.NewAction("scroll")
}

func (c *flowContent) handleScroll(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		elem := app.Window().GetElementByID("page-end-anchor")
		boundary := elem.JSValue().Call("getBoundingClientRect")
		bottom := boundary.Get("bottom").Int()

		_, height := app.Window().Size()

		if bottom-height < 0 && !c.paginationEnd {
			ctx.Dispatch(func(ctx app.Context) {
				c.pageNo++
				log.Println("new content page request fired")
			})
			return
		}
	})
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

		if _, ok := litterAPI("DELETE", "/api/flow", interactedPost, c.user.Nickname); !ok {
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
		if _, ok := litterAPI("PUT", "/api/flow", interactedPost, c.user.Nickname); !ok {
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
	ctx.Handle("delete", c.handleDelete)
	ctx.Handle("scroll", c.handleScroll)
	ctx.Handle("star", c.handleStar)

	c.paginationEnd = false
	c.pagination = 0
	c.pageNo = 1

	c.eventListener = app.Window().AddEventListener("scroll", c.onScroll)
}

func (c *flowContent) OnDismount() {
	// https://go-app.dev/reference#BrowserWindow
	//c.eventListener()
}

func (c *flowContent) OnNav(ctx app.Context) {
	c.loaderShow = true
	c.loaderShowImage = true

	toastText := ""

	ctx.Async(func() {
		var enUser string
		var user models.User

		ctx.LocalStorage().Get("user", &enUser)

		// decode, decrypt and unmarshal the local storage string
		if err := prepare(enUser, &user); err != nil {
			toastText = "frontend decoding/decryption failed: " + err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		postsRaw := struct {
			Posts map[string]models.Post `json:"posts"`
		}{}

		// call the flow API endpoint to fetch all posts
		if byteData, _ := litterAPI("GET", "/api/flow", nil, user.Nickname); byteData != nil {
			err := json.Unmarshal(*byteData, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parsing error: " + err.Error()
			}
		} else {
			log.Println("cannot fetch post flow list")
			toastText = "API error: cannot fetch the post list"

		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.loggedUser = user.Nickname
			c.user = user

			c.pagination = 10
			c.pageNo = 1

			c.posts = postsRaw.Posts
			c.toastText = toastText
			c.toastShow = (toastText != "")

			c.loaderShow = false
			c.loaderShowImage = false
		})
		return
	})
}

func (c *flowContent) Render() app.UI {
	var sortedPosts []models.Post

	// fetch posts and put them in an array
	for _, sortedPost := range c.posts {
		// do not append a post that is not meant to be shown
		if !c.user.FlowList[sortedPost.Nickname] && sortedPost.Nickname != "system" {
			continue
		}

		sortedPosts = append(sortedPosts, sortedPost)
	}

	// order posts by timestamp DESC
	sort.SliceStable(sortedPosts, func(i, j int) bool {
		return sortedPosts[i].Timestamp.After(sortedPosts[j].Timestamp)
	})

	// prepare posts according to the actual pagination and pageNo
	pagedPosts := []models.Post{}

	end := len(sortedPosts)
	start := 0

	stop := func(c *flowContent) int {
		var pos int

		if c.pagination > 0 {
			// (c.pageNo - 1) * c.pagination + c.pagination
			pos = c.pageNo * c.pagination
		}

		if pos > end {
			// kill the eventListener (observers scrolling)
			c.eventListener()
			c.paginationEnd = true

			return (end - 1)
		}

		if pos < 0 {
			return 0
		}

		return pos
	}(c)

	if end > 0 && stop > 0 {
		pagedPosts = sortedPosts[start:stop]
	}

	return app.Main().Class("responsive").Body(
		app.H5().Text("littr flow").Style("padding-top", config.HeaderTopPadding),
		app.P().Text("exclusive content incoming frfr"),
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.dismissToast).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// flow posts/articles
		app.Table().Class("border left-align").ID("table-flow").Body(
			// table header
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("align-left").Text("nickname, content, timestamp"),
				),
			),

			// table body
			app.TBody().Body(
				//app.Range(c.posts).Map(func(key string) app.UI {
				app.Range(pagedPosts).Slice(func(idx int) app.UI {
					//post := c.sortedPosts[idx]
					post := sortedPosts[idx]
					key := post.ID

					// only show posts of users in one's flowList
					if !c.user.FlowList[post.Nickname] && post.Nickname != "system" {
						return nil
					}

					return app.Tr().Class().Body(
						app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).On("scroll", c.onScroll).Body(
							app.P().Body(
								app.B().Text(post.Nickname).Class("deep-orange-text"),
							),

							app.If(post.Type == "fig",
								app.Article().Class("medium no-padding transparent").Body(
									app.If(c.loaderShowImage,
										app.Div().Class("small-space"),
										app.Div().Class("loader center large deep-orange active"),
									),
									app.Img().Class("no-padding absolute center middle lazy").Src(post.Content).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
								),
							).Else(
								app.Article().Class("post").Style("max-width", "100%").Body(
									app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto"),
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
		app.Div().ID("page-end-anchor"),
		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.Div().Class("loader center large deep-orange active"),
		),
	)
}
