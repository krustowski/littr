package flow

import (
	"sort"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/frontend/atomic/organisms"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
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

	// order posts by timestamp DESC
	sort.SliceStable(sortedPosts, func(i, j int) bool {
		if c.singlePostID != "" {
			return sortedPosts[i].Timestamp.Before(sortedPosts[j].Timestamp)
		}

		return sortedPosts[i].Timestamp.After(sortedPosts[j].Timestamp)
	})

	return sortedPosts
}

func (c *Content) Render() app.UI {
	//counter := 0

	return app.Main().Class("responsive").Body(
		// Page heading
		&molecules.FlowHeader{
			SingleUser:      c.users[c.userFlowNick],
			SinglePostID:    c.singlePostID,
			Hashtag:         c.hashtag,
			ButtonsDisabled: c.buttonDisabled,
			RefreshClicked:  c.refreshClicked,
		},

		// SingleUser view (profile mode)
		&organisms.SingleUserSummary{
			LoggedUser:              c.user,
			SingleUser:              c.users[c.userFlowNick],
			ButtonsDisabled:         c.buttonDisabled,
			OnClickFollowActionName: "follow",
		},

		app.Div().Class("space"),

		// Post deletion modal.
		&organisms.ModalPostDelete{
			ModalShow:            c.deletePostModalShow,
			ModalButtonsDisabled: c.deleteModalButtonsDisabled,
			OnClickDismiss:       c.onClickDismiss,
			OnClickDelete:        c.onClickDelete,
		},

		// Post reply modal.
		&organisms.ModalPostReply{
			PostOriginal:         c.posts[c.interactedPostKey],
			ModalShow:            c.modalReplyActive,
			ModalButtonsDisabled: c.postButtonsDisabled,
			OnClickDismiss:       c.onClickDismiss,
			OnClickReply:         c.onClickPostReply,
			OnFigureUpload:       c.handleFigUpload,
		},

		// The very post feed.
		&organisms.PostFeed{
			Posts:               c.posts,
			Users:               c.users,
			LoaderShowImage:     c.loaderShowImage,
			ButtonsDisabled:     c.buttonDisabled,
			LoggedUserNickname:  c.user.Nickname,
			SortedPosts:         c.sortPosts(),
			OnClickImage:        c.onClickImage,
			OnClickStar:         c.onClickStar,
			OnClickReply:        c.onClickReply,
			OnClickLink:         c.onClickLink,
			OnClickDeleteButton: c.onClickDeleteButton,
			OnClickUser:         c.onClickUserFlow,
			OnMouseEnter:        c.onMouseEnter,
			OnMouseLeave:        c.onMouseLeave,
		},

		// flow posts/articles
		/*app.Table().Class("left-aligni border").ID("table-flow").Style("padding", "0 0 2em 0").Style("border-spacing", "0.1em").Body(
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

					systemLink := func() string {
						// A system post about a new poll.
						if post.PollID != "" {
							return "/polls/" + post.PollID
						}

						// A system post about a new user.
						if post.Nickname == "system" && post.Type == "user" {
							return "/flow/users/" + post.Figure
						}

						// A system post about a new poll (legacy).
						return "/polls"
					}()

					return app.Tr().Class().Class("bottom-padding").Body(
						// special system post
						app.If(post.Nickname == "system", func() app.UI {
							return app.Td().Class("align-left").Attr("touch-action", "none").Body(
								app.Article().Class("responsive border blue-border info thicc margin-top center-align").Body(
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
							)

							// other posts
						}).Else(func() app.UI {
							//app.Td().Class("post align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).On("scroll", c.onScroll).Body(
							return app.Td().Class("align-left").Attr("data-author", post.Nickname).Attr("data-timestamp", post.Timestamp.UnixNano()).Attr("touch-action", "none").Body(

								// post header (author avatar + name + link button)
								app.Div().Class("row top-padding").Body(
									app.Img().ID(post.Nickname).Title("user's avatar").Class("responsive max left").Src(c.users[post.Nickname].AvatarURL).Style("max-width", "60px").Style("border-radius", "50%").OnClick(c.onClickUserFlow),
									app.P().Class("max").Body(
										app.A().Title("user's flow link").Class("bold deep-orange-text").OnClick(c.onClickUserFlow).ID(post.Nickname).Body(
											app.Span().ID("user-flow-link").Class("large-text bold deep-orange-text").Text(post.Nickname).OnMouseEnter(c.onMouseEnter).OnMouseLeave(c.onMouseLeave),
										),
										//app.B().Text(post.Nickname).Class("deep-orange-text"),
									),
									app.Button().ID(key).Title("link to this post (to clipboard)").Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
										app.I().Text("link"),
									),
								),

								// pic post
								app.If(post.Type == "fig", func() app.UI {
									return app.Article().Style("z-index", "5").Class("transparent medium no-margin thicc").Body(
										app.If(c.loaderShowImage, func() app.UI {
											return app.Div().Body(
												app.Div().Class("small-space"),
												app.Div().Class("loader center large deep-orange active"),
											)
										}),
										//app.Img().Class("no-padding center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
										app.Img().Class("no-padding center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
									)

									// reply + post
								}).Else(func() app.UI {
									return app.Div().Body(
										app.If(post.ReplyToID != "", func() app.UI {
											return app.Article().Class("black-text border reply thicc").Style("max-width", "100%").Body(
												app.Div().Class("row max").Body(
													app.If(previousDetailsSummary != "", func() app.UI {
														return app.Details().Class("max").Body(
															app.Summary().Text(previousDetailsSummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
															app.Div().Class("space"),
															app.Span().Class("bold").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
														)
													}).Else(func() app.UI {
														return app.Span().Class("max bold").Text(previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line")
													}),

													app.Button().Title("link to original post").ID(post.ReplyToID).Class("transparent circle").OnClick(c.onClickLink).Disabled(c.buttonDisabled).Body(
														app.I().Text("history"),
													),
												),
											)
										}),

										app.If(len(post.Content) > 0, func() app.UI {
											return app.Article().Class("border thicc").Style("max-width", "100%").Body(
												app.If(postDetailsSummary != "", func() app.UI {
													return app.Details().Body(
														app.Summary().Text(postDetailsSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
														app.Div().Class("space"),
														app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
													)
												}).Else(func() app.UI {
													return app.Span().Text(post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line")
												}),
											)
										}),

										app.If(post.Figure != "", func() app.UI {
											return app.Article().Style("z-index", "4").Class("transparent medium thicc").Body(
												app.If(c.loaderShowImage, func() app.UI {
													return app.Div().Body(
														app.Div().Class("small-space"),
														app.Div().Class("loader center large deep-orange active"),
													)
												}),
												//app.Img().Class("no-padding center middle lazy").Src(pixDestination).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
												app.Img().Class("no-padding center middle lazy").Src(imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(c.onClickImage).ID(post.ID),
											)
										}),
									)
								}),

								// post footer (timestamp + reply buttom + star/delete button)
								app.Div().Class("row").Body(
									app.Div().Class("max").Body(
										//app.Text(post.Timestamp.Format("Jan 02, 2006 / 15:04:05")),
										app.Text(postTimestamp),
									),
									app.If(post.Nickname != "system", func() app.UI {
										return app.Div().Body(
											app.If(post.ReplyCount > 0, func() app.UI {
												return app.B().Title("reply count").Text(post.ReplyCount).Class("left-padding")
											}),
											app.Button().Title("reply").ID(key).Class("transparent circle").OnClick(c.onClickReply).Disabled(c.buttonDisabled).Body(
												app.I().Text("reply"),
											),
										)
									}),
									app.If(c.user.Nickname == post.Nickname, func() app.UI {
										return app.Div().Body(
											app.B().Title("reaction count").Text(post.ReactionCount).Class("left-padding"),
											//app.Button().ID(key).Class("transparent circle").OnClick(c.onClickDelete).Disabled(c.buttonDisabled).Body(
											app.Button().Title("delete this post").ID(key).Class("transparent circle").OnClick(c.onClickDeleteButton).Disabled(c.buttonDisabled).Body(
												app.I().Text("delete"),
											),
										)
									}).Else(func() app.UI {
										return app.Div().Body(
											app.B().Title("reaction count").Text(post.ReactionCount).Class("left-padding"),
											app.Button().Title("increase the reaction count").ID(key).Class("transparent circle").OnClick(c.onClickStar).Disabled(c.buttonDisabled).Attr("touch-action", "none").Body(
												//app.I().Text("bomb"),			// literal bomb
												//app.I().Text("nest_eco_leaf"),	// leaf
												app.I().Text("ac_unit"), // snowflake
											),
										)
									}),
								),
							)
						}),
					)
				}),
			),
		),*/

		&atoms.Loader{
			ID:         "page-end-anchor",
			ShowLoader: c.loaderShow,
		},
	)
}
