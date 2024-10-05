package flow

import (
	"net/url"
	"sort"
	"strings"
	"time"

	"go.vxn.dev/littr/configs"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) sortPosts() []models.Post {
	var sortedPosts []models.Post

	posts := c.posts
	if posts == nil {
		posts = make(map[string]models.Post)
	}

	flowList := c.user.FlowList
	if len(flowList) == 0 {
		return sortedPosts
	}

	// fetch posts and put them in an array
	for _, sortedPost := range posts {
		// do not append a post that is not meant to be shown
		if !c.user.FlowList[sortedPost.Nickname] && sortedPost.Nickname != "system" && sortedPost.Nickname != c.userFlowNick {
			continue
		}

		sortedPosts = append(sortedPosts, sortedPost)
	}

	return sortedPosts
}

func (c *Content) Render() app.UI {
	toastColor := ""

	switch c.toastType {
	case "success":
		toastColor = "green10"
		break

	case "info":
		toastColor = "blue10"
		break

	default:
		toastColor = "red10"
	}

	counter := 0

	sortedPosts := c.sortPosts()

	// order posts by timestamp DESC
	sort.SliceStable(sortedPosts, func(i, j int) bool {
		if c.singlePostID != "" {
			return sortedPosts[i].Timestamp.Before(sortedPosts[j].Timestamp)
		}

		return sortedPosts[i].Timestamp.After(sortedPosts[j].Timestamp)
	})

	// compose a summary of a long post to be replied to
	replySummary := ""
	if c.modalReplyActive && len(c.posts[c.interactedPostKey].Content) > configs.MaxPostLength {
		replySummary = c.posts[c.interactedPostKey].Content[:configs.MaxPostLength/10] + "- [...]"
	}

	return app.Main().Class("responsive").Body(
		// page heading
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.If(c.userFlowNick != "" && !c.isPost,
					app.H5().Body(
						app.Text(c.userFlowNick+"'s flow"),

						app.If(c.users[c.userFlowNick].Private,
							app.Span().Class("bold").Body(
								app.I().Text("lock"),
							),
						),
					),
				).ElseIf(c.singlePostID != "" && c.isPost,
					app.H5().Text("single post and replies"),
				).ElseIf(c.hashtag != "" && len(c.hashtag) < 20,
					app.H5().Text("hashtag #"+c.hashtag),
				).ElseIf(c.hashtag != "" && len(c.hashtag) >= 20,
					app.H5().Text("hashtag"),
				).Else(
					app.H5().Text("flow"),
					//app.P().Text("exclusive content incoming frfr"),
				),
			),

			app.Div().Class("small-padding").Body(
				app.Button().ID("refresh-button").Title("refresh flow [R]").Class("border black white-text bold").Style("border-radius", "8px").OnClick(c.onClickRefresh).Disabled(c.postButtonsDisabled).Body(
					app.If(c.refreshClicked,
						app.Progress().Class("circle deep-orange-border small"),
					),
					app.Text("refresh"),
				),
			),
		),

		app.If(c.userFlowNick != "" && !c.isPost,
			app.Div().Class("row top-padding").Body(
				app.Img().Class("responsive max left").Src(c.users[c.userFlowNick].AvatarURL).Style("max-width", "80px").Style("border-radius", "50%"),
				/*;app.P().Class("max").Body(
					app.A().Class("bold deep-orange-text").Text(c.singlePostID).ID(c.singlePostID),
					//app.B().Text(post.Nickname).Class("deep-orange-text"),
				),*/

				//app.If(c.users[c.userFlowNick].About != "",
				app.Article().Class("max").Style("word-break", "break-word").Style("hyphens", "auto").Text(c.users[c.userFlowNick].About),
				//),
				app.Button().ID(c.userFlowNick).Class("black border white-text").Style("border-radius", "8px").OnClick(c.onClickFollow).Disabled(c.buttonDisabled || c.userFlowNick == c.user.Nickname).Body(
					app.If(c.user.FlowList[c.userFlowNick],
						app.Span().Text("unfollow"),
					).Else(
						app.Span().Text("follow"),
					),
				),
			),
		),

		app.Div().Class("space"),

		// snackbar
		app.If(c.toastText != "",
			app.A().OnClick(c.onClickDismiss).Body(
				app.Div().ID("snackbar").Class("snackbar white-text top active "+toastColor).Body(
					app.I().Text("error"),
					app.Span().Text(c.toastText),
				),
			),
		),

		// post deletion modal
		app.If(c.deletePostModalShow,
			app.Dialog().ID("delete-modal").Class("grey9 white-text active").Style("border-radius", "8px").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("post deletion"),
				),

				app.Div().Class("space"),
				app.Article().Class("row").Body(
					app.I().Text("warning").Class("amber-text"),
					app.P().Class("max").Body(
						app.Span().Text("are you sure you want to delete your post?"),
					),
				),
				app.Div().Class("space"),

				app.Div().Class("row").Body(
					app.Button().Class("max border red10 white-text").Style("border-radius", "8px").OnClick(c.onClickDelete).Disabled(c.deleteModalButtonsDisabled).Body(
						app.If(c.deleteModalButtonsDisabled,
							app.Progress().Class("circle white-border small"),
						),
						app.Text("yeah"),
					),
					app.Button().Class("max border black white-text").Style("border-radius", "8px").Text("nope").OnClick(c.onClickDismiss).Disabled(c.deleteModalButtonsDisabled),
				),
			),
		),

		//app.Div().ID("overlay").Class("overlay").OnClick(c.onClickDismiss).Style("z-index", "50"),

		// sketchy reply modal
		app.If(c.modalReplyActive,
			app.Dialog().ID("reply-modal").Class("grey9 white-text center-align active").Style("max-width", "90%").Style("border-radius", "8px").Style("z-index", "75").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("reply"),
				),
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

				app.Div().Class("field label textarea border extra deep-orange-text").Body(
					//app.Textarea().Class("active").Name("replyPost").OnChange(c.ValueTo(&c.replyPostContent)).AutoFocus(true).Placeholder("reply to: "+c.posts[c.interactedPostKey].Nickname),
					app.Textarea().Class("active").Name("replyPost").Text(c.replyPostContent).OnChange(c.ValueTo(&c.replyPostContent)).AutoFocus(true).ID("reply-textarea"),
					app.Label().Text("reply to: "+c.posts[c.interactedPostKey].Nickname).Class("active deep-orange-text"),
					//app.Label().Text("text").Class("active"),
				),
				app.Div().Class("field label border extra deep-orange-text").Body(
					app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload).Accept("image/*"),
					app.Input().Class("active").Type("text").Value(c.newFigFile).Disabled(true),
					app.Label().Text("image").Class("active deep-orange-text"),
					app.I().Text("image"),
				),

				app.Div().Class("row").Body(
					app.Button().Class("max border deep-orange7 white-text bold").Text("cancel").Style("border-radius", "8px").OnClick(c.onClickDismiss).Disabled(c.postButtonsDisabled),
					app.Button().ID("reply").Class("max border deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClickPostReply).Disabled(c.postButtonsDisabled).Body(
						app.If(c.postButtonsDisabled,
							app.Progress().Class("circle white-border small"),
						),
						app.Text("reply"),
					),
				),
				app.Div().Class("space"),
			),
		),

		// flow posts/articles
		app.Table().Class("left-aligni border").ID("table-flow").Style("padding", "0 0 2em 0").Style("border-spacing", "0.1em").Body(
			// table body
			app.TBody().Body(
				//app.Range(c.posts).Map(func(key string) app.UI {
				//app.Range(pagedPosts).Slice(func(idx int) app.UI {
				app.Range(sortedPosts).Slice(func(idx int) app.UI {
					counter++
					if counter > c.pagination*c.pageNo {
						return nil
					}

					//post := c.sortedPosts[idx]
					post := sortedPosts[idx]
					key := post.ID

					previousContent := ""

					// prepare reply parameters to render
					if post.ReplyToID != "" {
						if c.hideReplies {
							return nil
						}

						if previous, found := c.posts[post.ReplyToID]; found {
							if value, foundU := c.user.FlowList[previous.Nickname]; (!value || !foundU) && c.users[previous.Nickname].Private {
								previousContent = "this content is private"
							} else {
								previousContent = previous.Nickname + " posted: " + previous.Content
							}
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

					if c.userFlowNick != "" {
						if post.Nickname != c.userFlowNick {
							return nil
						}
					}

					// only show posts of users in one's flowList
					if !c.user.FlowList[post.Nickname] && post.Nickname != "system" {
						return nil
					}

					// check the post's length, on threshold use <details> tag
					postDetailsSummary := ""
					if len(post.Content) > configs.MaxPostLength {
						postDetailsSummary = post.Content[:configs.MaxPostLength/10] + "- [...]"
					}

					// the same as above with the previous post's length for reply render
					previousDetailsSummary := ""
					if len(previousContent) > configs.MaxPostLength {
						previousDetailsSummary = previousContent[:configs.MaxPostLength/10] + "- [...]"
					}

					// fetch the image
					var imgSrc string

					// check the URL/URI format
					if post.Type == "fig" {
						if _, err := url.ParseRequestURI(post.Content); err == nil {
							imgSrc = post.Content
						} else {
							fileExplode := strings.Split(post.Content, ".")
							extension := fileExplode[len(fileExplode)-1]

							imgSrc = "/web/pix/thumb_" + post.Content
							if extension == "gif" {
								imgSrc = "/web/click-to-see-gif.jpg"
							}
						}
					} else if post.Type == "post" {
						if _, err := url.ParseRequestURI(post.Figure); err == nil {
							imgSrc = post.Figure
						} else {
							fileExplode := strings.Split(post.Figure, ".")
							extension := fileExplode[len(fileExplode)-1]

							imgSrc = "/web/pix/thumb_" + post.Figure
							if extension == "gif" {
								imgSrc = "/web/click-to-see.gif"
							}
						}
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

						if resp, ok = littrAPI("POST", "/api/pix", payload, c.user.Nickname); !ok {
							log.Println("api failed")
							imgSrc = "/web/android-chrome-512x512.png"
						} else {
							imgSrc = "data:image/*;base64," + b64.StdEncoding.EncodeToString(*resp)
						}
					}*/

					var postTimestamp string

					// use JS toLocaleString() function to reformat the timestamp
					// use negated LocalTimeMode boolean as true! (HELP)
					if !c.user.LocalTimeMode {
						postLocale := app.Window().
							Get("Date").
							New(post.Timestamp.Format(time.RFC3339))

						postTimestamp = postLocale.Call("toLocaleString", "en-GB").String()
					} else {
						postTimestamp = post.Timestamp.Format("Jan 02, 2006 / 15:04:05")
					}

					// omit older system messages for new users
					if post.Nickname == "system" && post.Timestamp.Before(c.user.RegisteredTime) {
						return nil
					}

					systemLink := "/polls"
					if post.Nickname == "system" && post.Type == "user" {
						systemLink = "/flow/user/" + post.Figure
					}

					return app.Tr().Class().Class("bottom-padding").Body(
						// special system post
						app.If(post.Nickname == "system",
							app.Td().Class("post align-left").Attr("touch-action", "none").Body(
								app.Article().Class("responsive margin-top center-align").Body(
									app.A().Href(systemLink).Body(
										app.Span().Class("bold").Text(post.Content),
									),
								),
								app.Div().Class("row").Body(
									app.Div().Class("max").Body(
										//app.Text(post.Timestamp.Format("Jan 02, 2006 / 15:04:05")),
										app.Text(postTimestamp),
									),
								),
							),

						// other posts
						).Else(
							//app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).On("scroll", c.onScroll).Body(
							app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).Attr("touch-action", "none").Body(

								// post header (author avatar + name + link button)
								app.Div().Class("row top-padding").Body(
									app.Img().Title("user's avatar").Class("responsive max left").Src(c.users[post.Nickname].AvatarURL).Style("max-width", "60px").Style("border-radius", "50%"),
									app.P().Class("max").Body(
										app.A().Title("user's flow link").Class("bold deep-orange-text").OnClick(c.onClickUserFlow).ID(post.Nickname).Body(
											app.Span().Class("large-text bold deep-orange-text").Text(post.Nickname),
										),
										//app.B().Text(post.Nickname).Class("deep-orange-text"),
									),
									app.Button().ID(key).Title("link to this post (to clipboard)").Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
										app.I().Text("link"),
									),
								),

								// pic post
								app.If(post.Type == "fig",
									app.Article().Style("z-index", "5").Style("border-radius", "8px").Class("transparent medium no-margin").Body(
										app.If(c.loaderShowImage,
											app.Div().Class("small-space"),
											app.Div().Class("loader center large deep-orange active"),
										),
										//app.Img().Class("no-padding center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
										app.Img().Class("no-padding center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
									),

								// reply + post
								).Else(
									app.If(post.ReplyToID != "",
										app.Article().Class("black-text yellow10").Style("border-radius", "8px").Style("max-width", "100%").Body(
											app.Div().Class("row max").Body(
												app.If(previousDetailsSummary != "",
													app.Details().Class("max").Body(
														app.Summary().Text(previousDetailsSummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
														app.Div().Class("space"),
														app.Span().Class("bold").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
													),
												).Else(
													app.Span().Class("max bold").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
												),

												app.Button().Title("link to original post").ID(post.ReplyToID).Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
													app.I().Text("history"),
												),
											),
										),
									),

									app.If(len(post.Content) > 0,
										app.Article().Class("surface-container-highest").Style("border-radius", "8px").Style("max-width", "100%").Body(
											app.If(postDetailsSummary != "",
												app.Details().Body(
													app.Summary().Text(postDetailsSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
													app.Div().Class("space"),
													app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
												),
											).Else(
												app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
											),
										),
									),

									app.If(post.Figure != "",
										app.Article().Style("z-index", "4").Style("border-radius", "8px").Class("transparent medium medium").Body(
											app.If(c.loaderShowImage,
												app.Div().Class("small-space"),
												app.Div().Class("loader center large deep-orange active"),
											),
											//app.Img().Class("no-padding center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
											app.Img().Class("no-padding center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
										),
									),
								),

								// post footer (timestamp + reply buttom + star/delete button)
								app.Div().Class("row").Body(
									app.Div().Class("max").Body(
										//app.Text(post.Timestamp.Format("Jan 02, 2006 / 15:04:05")),
										app.Text(postTimestamp),
									),
									app.If(post.Nickname != "system",
										app.If(post.ReplyCount > 0,
											app.B().Title("reply count").Text(post.ReplyCount).Class("left-padding"),
										),
										app.Button().Title("reply").ID(key).Class("transparent circle").OnClick(c.onClickReply).Disabled(c.buttonDisabled).Body(
											app.I().Text("reply"),
										),
									),
									app.If(c.user.Nickname == post.Nickname,
										app.B().Title("reaction count").Text(post.ReactionCount).Class("left-padding"),
										//app.Button().ID(key).Class("transparent circle").OnClick(c.onClickDelete).Disabled(c.buttonDisabled).Body(
										app.Button().Title("delete this post").ID(key).Class("transparent circle").OnClick(c.onClickDeleteButton).Disabled(c.buttonDisabled).Body(
											app.I().Text("delete"),
										),
									).Else(
										app.B().Title("reaction count").Text(post.ReactionCount).Class("left-padding"),
										app.Button().Title("increase the reaction count").ID(key).Class("transparent circle").OnClick(c.onClickStar).Disabled(c.buttonDisabled).Attr("touch-action", "none").Body(
											//app.I().Text("ac_unit"),
											//app.I().Text("bomb"),
											app.I().Text("nest_eco_leaf"),
										),
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
