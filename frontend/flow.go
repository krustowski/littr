package frontend

import (
	//b64 "encoding/base64"
	"encoding/json"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

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

	buttonDisabled      bool
	postButtonsDisabled bool
	modalReplyActive    bool
	replyPostContent    string

	interactedPostKey string
	singlePostID      string
	isPost            bool

	paginationEnd bool
	pagination    int
	pageNo        int

	postKey     string
	posts       map[string]models.Post
	users       map[string]models.User
	sortedPosts []models.Post

	refreshClicked bool
}

func (p *FlowPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("flow / littr")
}

func (p *FlowPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&flowContent{},
	)
}

func (c *flowContent) onClickLink(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host

	// write the link to browsers's clipboard
	app.Window().Get("navigator").Get("clipboard").Call("writeText", scheme+"://"+host+"/flow/"+key)
	ctx.Navigate("/flow/" + key)
}

func (c *flowContent) onClickDismiss(ctx app.Context, e app.Event) {
	c.toastShow = false
	c.toastText = ""
	c.modalReplyActive = false
	c.buttonDisabled = false
}

func (c *flowContent) onClickImage(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	src := ctx.JSSrc().Get("src").String()

	split := strings.Split(src, ".")
	ext := split[len(split)-1]

	// image preview (thumbnail) to the actual image logic
	if strings.Contains(src, "thumb") {
		ctx.JSSrc().Set("src", "/web/pix/"+key+"."+ext)
		//ctx.JSSrc().Set("style", "max-height: 90vh; max-height: 100%; transition: max-height 0.1s; z-index: 1; max-width: 100%; background-position: center")
		ctx.JSSrc().Set("style", "max-height: 90vh; transition: max-height 0.1s; z-index: 1; max-width: 100%; background-position")
	} else {
		ctx.JSSrc().Set("src", "/web/pix/thumb_"+key+"."+ext)
		ctx.JSSrc().Set("style", "z-index: 0; max-height: 100%; max-width: 100%")
	}
}

func (c *flowContent) handleImage(ctx app.Context, a app.Action) {
	ctx.JSSrc().Set("src", "")
}

func (c *flowContent) onClickReply(ctx app.Context, e app.Event) {
	c.interactedPostKey = ctx.JSSrc().Get("id").String()

	c.modalReplyActive = true
	c.buttonDisabled = true
}

func (c *flowContent) onClickPostReply(ctx app.Context, e app.Event) {
	//c.interactedPostKey = ctx.JSSrc().Get("id").String()

	c.modalReplyActive = true
	c.postButtonsDisabled = true
	c.buttonDisabled = true

	ctx.NewAction("reply")
}

func (c *flowContent) handleReply(ctx app.Context, a app.Action) {
	ctx.Async(func() {
		toastText := ""

		// TODO: allow figs in replies
		// check if the contents is a valid URL, then change the type to "fig"
		postType := "post"

		// trim the spaces on the extremites
		replyPost := strings.TrimSpace(c.replyPostContent)

		if replyPost == "" {
			toastText = "no valid reply entered"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		//newPostID := time.Now()
		//stringID := strconv.FormatInt(newPostID.UnixNano(), 10)

		path := "/api/flow"

		// TODO: the Post data model has to be changed
		// migrate Post.ReplyID (int) to Post.ReplyID (string)
		// ReplyID is to be string key to easily refer to other post
		payload := models.Post{
			//ID:        stringID,
			Nickname: c.user.Nickname,
			Type:     postType,
			Content:  replyPost,
			//Timestamp: newPostID,
			//ReplyTo: replyID, <--- is type int
			ReplyToID: c.interactedPostKey,
		}

		postsRaw := struct {
			Posts map[string]models.Post `posts`
		}{}

		// add new post/poll to backend struct
		if resp, _ := litterAPI("POST", path, payload, c.user.Nickname); resp != nil {
			err := json.Unmarshal(*resp, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parsing error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			log.Println("cannot fetch post flow list")
			toastText = "API error: cannot fetch the post list"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		payloadNotif := struct {
			OriginalPost string `json:"original_post"`
		}{
			OriginalPost: c.interactedPostKey,
		}

		// create a notification
		if _, ok := litterAPI("PUT", "/api/push", payloadNotif, c.user.Nickname); !ok {
			toastText = "cannot PUT new notification"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			// add new post to post list on frontend side to render
			//c.posts[stringID] = payload
			c.posts = postsRaw.Posts

			c.modalReplyActive = false
			c.postButtonsDisabled = false
			c.buttonDisabled = false
		})
	})
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
	ctx.Handle("image", c.handleImage)
	ctx.Handle("reply", c.handleReply)
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

func (c *flowContent) onClickRefresh(ctx app.Context, e app.Event) {
	c.refreshClicked = true
	c.fetchUsers(ctx)
	c.fetchPosts(ctx)
}

func (c *flowContent) fetchPosts(ctx app.Context) {
	var toastText string

	postsRaw := struct {
		Posts map[string]models.Post `json:"posts"`
	}{}

	// call the flow API endpoint to fetch all posts
	if byteData, _ := litterAPI("GET", "/api/flow", nil, c.user.Nickname); byteData != nil {
		err := json.Unmarshal(*byteData, &postsRaw)
		if err != nil {
			log.Println(err.Error())
			toastText = "JSON parsing error: " + err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}
	} else {
		log.Println("cannot fetch post flow list")
		toastText = "API error: cannot fetch the post list"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.posts = postsRaw.Posts
		c.refreshClicked = false
	})
}

func (c *flowContent) fetchUsers(ctx app.Context) {
	var toastText string

	usersRaw := struct {
		Users map[string]models.User `json:"users"`
	}{}

	// call the flow API endpoint to fetch all users
	if byteData, _ := litterAPI("GET", "/api/users", nil, c.user.Nickname); byteData != nil {
		err := json.Unmarshal(*byteData, &usersRaw)
		if err != nil {
			log.Println(err.Error())
			toastText = "JSON parsing error: " + err.Error()

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}
	} else {
		log.Println("cannot fetch users list")
		toastText = "API error: cannot fetch the users list"

		ctx.Dispatch(func(ctx app.Context) {
			c.toastText = toastText
			c.toastShow = (toastText != "")
		})
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.users = usersRaw.Users
	})
}

func (c *flowContent) OnNav(ctx app.Context) {
	c.loaderShow = true
	c.loaderShowImage = true

	toastText := ""

	singlePostID := ""
	isPost := true
	url := strings.Split(ctx.Page().URL().Path, "/")

	if len(url) > 2 && url[2] != "" {
		singlePostID = url[2]
	}

	ctx.Async(func() {
		if _, err := strconv.Atoi(singlePostID); singlePostID != "" && err != nil {
			// prolly not a post ID, but an user's nickname
			isPost = false
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.isPost = isPost
		})

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

		usersRaw := struct {
			Users map[string]models.User `json:"users"`
		}{}

		// call the flow API endpoint to fetch all posts
		if byteData, _ := litterAPI("GET", "/api/flow", nil, user.Nickname); byteData != nil {
			err := json.Unmarshal(*byteData, &postsRaw)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parsing error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			log.Println("cannot fetch post flow list")
			toastText = "API error: cannot fetch the post list"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// call the flow API endpoint to fetch all users
		if byteData, _ := litterAPI("GET", "/api/users", nil, user.Nickname); byteData != nil {
			err := json.Unmarshal(*byteData, &usersRaw)
			if err != nil {
				log.Println(err.Error())
				toastText = "JSON parsing error: " + err.Error()

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = toastText
					c.toastShow = (toastText != "")
				})
				return
			}
		} else {
			log.Println("cannot fetch users list")
			toastText = "API error: cannot fetch the users list"

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		// try the singlePostID var if present
		if singlePostID != "" {
			_, found := postsRaw.Posts[singlePostID]
			if isPost && !found {
				toastText = "post not found"
			}

			if !isPost {
				toastText = ""

				if _, found = usersRaw.Users[singlePostID]; !found {
					toastText = "user not found"
				}
			}
		}

		// Storing HTTP response in component field:
		ctx.Dispatch(func(ctx app.Context) {
			c.loggedUser = user.Nickname
			c.user = user

			c.pagination = 10
			c.pageNo = 1

			c.users = usersRaw.Users
			c.posts = postsRaw.Posts
			c.singlePostID = singlePostID

			c.toastText = toastText
			c.toastShow = (toastText != "")

			c.loaderShow = false
			c.loaderShowImage = false
		})
		return
	})
}

func (c *flowContent) sortPosts() []models.Post {
	var sortedPosts []models.Post
	//var found bool

	posts := make(map[string]models.Post)

	/*if c.singlePostID != "" {
		if posts[c.singlePostID], found = c.posts[c.singlePostID]; !found {
			posts = c.posts
		} else {

		}
	} else {
		posts = c.posts
	}*/
	posts = c.posts

	// fetch posts and put them in an array
	for _, sortedPost := range posts {
		// do not append a post that is not meant to be shown
		if !c.user.FlowList[sortedPost.Nickname] && sortedPost.Nickname != "system" {
			continue
		}

		sortedPosts = append(sortedPosts, sortedPost)
	}

	return sortedPosts
}

func (c *flowContent) Render() app.UI {
	sortedPosts := c.sortPosts()

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

			return (end)
		}

		if pos < 0 {
			return 0
		}

		return pos
	}(c)

	if end > 0 && stop > 0 {
		pagedPosts = sortedPosts[start:stop]
	}

	// disable pagination for single posts (resource distress???)
	if c.singlePostID != "" {
		pagedPosts = sortedPosts[start:end]
	}

	// compose a summary of a long post to be replied to
	replySummary := ""
	if c.modalReplyActive && len(c.posts[c.interactedPostKey].Content) > config.MaxPostLength {
		replySummary = c.posts[c.interactedPostKey].Content[:config.MaxPostLength/10] + "- [...]"
	}

	return app.Main().Class("responsive").Body(
		// page heading
		app.Div().Class("row").Body(
			app.Div().Class("max").Body(
				app.H5().Text("littr flow").Style("padding-top", config.HeaderTopPadding),
				app.P().Text("exclusive content incoming frfr"),
			),
			app.Button().Class("border black white-text bold").OnClick(c.onClickRefresh).Disabled(c.postButtonsDisabled).Body(
				app.If(c.refreshClicked,
					app.Progress().Class("circle deep-orange-border small"),
				),
				app.Text("refresh"),
			),
		),
		app.Div().Class("space"),

		// snackbar
		app.A().OnClick(c.onClickDismiss).Body(
			app.If(c.toastText != "",
				app.Div().Class("snackbar red10 white-text top active").Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// sketchy reply modal
		app.If(c.modalReplyActive,
			app.Dialog().Class("grey9 white-text center-align active").Style("max-width", "90%").Body(
				app.Div().Class("space"),

				app.Article().Class("post").Style("max-width", "100%").Body(
					app.If(replySummary != "",
						app.Details().Body(
							app.Summary().Text(replySummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
							app.Div().Class("space"),
							app.Span().Text(c.posts[c.interactedPostKey].Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic"),
						),
					).Else(
						app.Span().Text(c.posts[c.interactedPostKey].Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic"),
					),
				),

				app.Div().Class("field textarea label border invalid extra deep-orange-text").Body(
					app.Textarea().Class("active").Name("replyPost").OnChange(c.ValueTo(&c.replyPostContent)).AutoFocus(true),
					app.Label().Text("reply to: "+c.posts[c.interactedPostKey].Nickname).Class("active"),
				),

				app.Nav().Class("center-align").Body(
					app.Button().Class("border deep-orange7 white-text bold").Text("cancel").OnClick(c.onClickDismiss).Disabled(c.postButtonsDisabled),
					app.Button().ID("").Class("border deep-orange7 white-text bold").Text("reply").OnClick(c.onClickPostReply).Disabled(c.postButtonsDisabled),
				),
				app.Div().Class("space"),
			),
		),

		// flow posts/articles
		app.Table().Class("border left-align").ID("table-flow").Body(
			// table body
			app.TBody().Body(
				//app.Range(c.posts).Map(func(key string) app.UI {
				app.Range(pagedPosts).Slice(func(idx int) app.UI {
					//post := c.sortedPosts[idx]
					post := sortedPosts[idx]
					key := post.ID

					previousContent := ""

					// prepare reply parameters to render
					if post.ReplyToID != "" {
						if previous, found := c.posts[post.ReplyToID]; found {
							previousContent = previous.Nickname + " posted: " + previous.Content
						} else {
							previousContent = "the post was deleted bye"
						}
					}

					// filter out not-single-post items
					if c.singlePostID != "" {
						if c.isPost && post.ID != c.singlePostID && c.singlePostID != post.ReplyToID {
							return nil
						}

						if _, found := c.users[c.singlePostID]; (!c.isPost && !found) || (found && post.Nickname != c.singlePostID) {
							return nil
						}
					}

					// only show posts of users in one's flowList
					if !c.user.FlowList[post.Nickname] && post.Nickname != "system" {
						return nil
					}

					// check the post's lenght, on threshold use <details> tag
					postDetailsSummary := ""
					if len(post.Content) > config.MaxPostLength {
						postDetailsSummary = post.Content[:config.MaxPostLength/10] + "- [...]"
					}

					// the same as above with the previous post's length for reply render
					previousDetailsSummary := ""
					if len(previousContent) > config.MaxPostLength {
						previousDetailsSummary = previousContent[:config.MaxPostLength/10] + "- [...]"
					}

					// fetch the image
					var imgSrc string

					// check the URL/URI format
					if _, err := url.ParseRequestURI(post.Content); err == nil {
						imgSrc = post.Content
					} else {
						imgSrc = "/web/pix/thumb_" + post.Content
					}

					// fetch binary image data
					/*if post.Type == "fig" && imgSrc == "" {
						payload := struct {
							PostID  string `json:"post_id"`
							Content string `json:"content"`
						}{
							PostID:  post.ID,
							Content: post.Content,
						}

						var resp *[]byte
						var ok bool

						if resp, ok = litterAPI("POST", "/api/pix", payload, c.user.Nickname); !ok {
							log.Println("api failed")
							imgSrc = "/web/android-chrome-512x512.png"
						} else {
							imgSrc = "data:image/*;base64," + b64.StdEncoding.EncodeToString(*resp)
						}
					}*/

					return app.Tr().Class().Body(
						//app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).On("scroll", c.onScroll).Body(
						app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).Body(

							// post header (author avatar + name + link button)
							app.Div().Class("row").Body(
								app.Img().Class("responsive max left").Src(c.users[post.Nickname].AvatarURL).Style("max-width", "60px").Style("border-radius", "50%"),
								app.P().Class("max").Body(
									app.B().Text(post.Nickname).Class("deep-orange-text"),
								),

								app.If(post.Nickname != "system",
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
										app.I().Text("link"),
									),
								),
							),

							// pic post
							app.If(post.Type == "fig",
								app.Article().Class("medium no-padding transparent").Body(
									app.If(c.loaderShowImage,
										app.Div().Class("small-space"),
										app.Div().Class("loader center large deep-orange active"),
									),
									//app.Img().Class("no-padding absolute center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
									app.Img().Class("no-padding absolute center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
								),

							// reply + post
							).Else(
								app.If(post.ReplyToID != "",
									app.Article().Class("post black-text yellow10").Style("max-width", "100%").Body(
										app.Div().Class("row max").Body(
											app.If(previousDetailsSummary != "",
												app.Details().Class("max").Body(
													app.Summary().Text(previousDetailsSummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
													app.Div().Class("space"),
													app.Span().Class("italic").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto"),
												),
											).Else(
												app.Span().Class("max italic").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto"),
											),

											app.Button().ID(post.ReplyToID).Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
												app.I().Text("history"),
											),
										),
									),
								),
								app.Article().Class("post").Style("max-width", "100%").Body(
									app.If(postDetailsSummary != "",
										app.Details().Body(
											app.Summary().Text(postDetailsSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
											app.Div().Class("space"),
											app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto"),
										),
									).Else(
										app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto"),
									),
								),
							),

							// post footer (timestamp + reply buttom + star/delete button)
							app.Div().Class("row").Body(
								app.Div().Class("max").Body(
									app.Text(post.Timestamp.Format("Jan 02, 2006; 15:04:05 -0700")),
								),
								app.If(post.Nickname != "system",
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickReply).Disabled(c.buttonDisabled).Body(
										app.I().Text("reply"),
									),
								),
								app.If(c.user.Nickname == post.Nickname,
									app.B().Text(post.ReactionCount),
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickDelete).Disabled(c.buttonDisabled).Body(
										app.I().Text("delete"),
									),
								).Else(
									app.B().Text(post.ReactionCount),
									app.Button().ID(key).Class("transparent circle").OnClick(c.onClickStar).Disabled(c.buttonDisabled).Body(
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
			app.Progress().Class("circle center large deep-orange-border active"),
		),
	)
}
