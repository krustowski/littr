package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"
)

type UserFeed struct {
	app.Compo

	LoggedUser models.User

	SortedUsers []models.User
	Users       map[string]models.User

	FlowStats map[string]int
	UserStats map[string]models.UserStat

	Pagination int
	PageNo     int

	ButtonsDisabled bool
	LoaderShowImage bool

	OnClickUserActionName     string
	OnClickNicknameActionName string
	OnMouseEnterActionName    string
	OnMouseLeaveActionName    string

	// Per user vars. Modified via the processUser() function.
	isInFlow    bool
	isShaded    bool
	isRequested bool
}

func (u *UserFeed) processUser(user models.User) bool {
	u.isInFlow = false
	u.isShaded = false
	u.isRequested = false

	var found bool

	if u.LoggedUser.FlowList != nil {
		if u.isInFlow, found = u.LoggedUser.FlowList[user.Nickname]; found && u.isInFlow {
			u.isInFlow = true
		}
	}

	if u.LoggedUser.ShadeList != nil {
		if u.isShaded, found = u.LoggedUser.ShadeList[user.Nickname]; found && u.isShaded {
			u.isShaded = true
		}
	}

	if user.RequestList != nil {
		if u.isRequested, found = user.RequestList[u.LoggedUser.Nickname]; !found {
			u.isRequested = false
		}
	}

	if !user.Searched || user.Nickname == "system" {
		return false
	}

	return true
}

func (u *UserFeed) Render() app.UI {
	return app.Div().Class("post-feed").Body(
		app.Range(u.SortedUsers).Slice(func(idx int) app.UI {
			user := u.SortedUsers[idx]

			if !u.processUser(user) {
				return nil
			}

			return app.Div().Class("post").Body(
				app.Div().Class("row medium top-padding").Body(
					//app.Img().ID(user.Nickname).Class("responsive max left").Src(user.AvatarURL).Style("max-width", "60px").Style("border-radius", "50%").OnClick(c.onClickUser),
					&atoms.Image{
						ID:                user.Nickname,
						Class:             "responsive max left",
						Src:               user.AvatarURL,
						Styles:            map[string]string{"max-width": "60px", "border-radius": "50%"},
						OnClickActionName: u.OnClickUserActionName,
					},

					app.If(user.Private, func() app.UI {
						return &atoms.UserNickname{
							Class:                  "deep-orange-text bold large-text",
							Icon:                   "lock",
							Nickname:               user.Nickname,
							SpanID:                 user.Nickname,
							Title:                  "user's nickname",
							Text:                   user.Nickname,
							OnClickActionName:      u.OnClickNicknameActionName,
							OnMouseEnterActionName: u.OnMouseEnterActionName,
							OnMouseLeaveActionName: u.OnMouseLeaveActionName,
						}
					}).Else(func() app.UI {
						return &atoms.UserNickname{
							Class:                  "deep-orange-text bold large-text",
							Nickname:               user.Nickname,
							SpanID:                 user.Nickname,
							Title:                  "user's nickname",
							Text:                   user.Nickname,
							OnClickActionName:      u.OnClickNicknameActionName,
							OnMouseEnterActionName: u.OnMouseEnterActionName,
							OnMouseLeaveActionName: u.OnMouseLeaveActionName,
						}
					}),

					// User's stats --- flower count
					&molecules.Counter{
						Count:             u.UserStats[user.Nickname].FlowerCount,
						ID:                user.Nickname,
						Title:             "flower count",
						Icon:              "group",
						OnClickActionName: u.OnClickNicknameActionName,
					},

					&molecules.Counter{
						Count:             u.UserStats[user.Nickname].FlowerCount,
						ID:                user.Nickname,
						Title:             "flower count",
						Icon:              "group",
						OnClickActionName: u.OnClickNicknameActionName,
					},
				),
			)
		}),
	)
}

/*
	// users table
	return app.Table().Class("border").ID("table-users").Style("width", "100%").Style("border-spacing", "0.1em").Style("padding", "0 0 2em 0").Body(
		app.TBody().Body(
			app.Range(pagedUsers).Slice(func(idx int) app.UI {
				return app.Tr().Body(
					app.Td().Class("left-align").Body(

						// cell's header
						app.Div().Class("row medium top-padding").Body(
							app.Img().ID(user.Nickname).Class("responsive max left").Src(user.AvatarURL).Style("max-width", "60px").Style("border-radius", "50%").OnClick(c.onClickUser),

							app.If(user.Private, func() app.UI {
								return app.Div().Body(
									// nasty hack to ensure the padding lock icon is next to nickname
									app.P().ID(user.Nickname).Class("deep-orange-text bold").OnClick(c.onClickUser).Body(
										app.Span().Class("large-text bold deep-orange-text").Text(user.Nickname),
									),

									// show private mode
									app.Span().Class("bold max").Body(
										app.I().Text("lock"),
									),
								)
							}).Else(func() app.UI {
								return app.P().ID(user.Nickname).Class("deep-orange-text bold max").OnClick(c.onClickUser).Body(
									app.Span().Class("large-text bold deep-orange-text").Text(user.Nickname),
								)
							}),

							// user's stats --- flower count
							app.B().Title("flower count").Text(c.userStats[user.Nickname].FlowerCount).Class("left-padding"),
							app.Span().Title("flower count").Class("bold").Body(
								//app.I().Text("filter_vintage"),
								app.I().Text("group"),
							),

							// user's stats --- post count
							app.B().Title("post count").Text(c.userStats[user.Nickname].PostCount).Class("left-padding"),
							app.Span().Title("post count (link to their flow)").Class("bold").OnClick(c.onClickUserFlow).ID(user.Nickname).Body(
								app.I().Text("news"),
							),

							// more button
							/*
								app.If(shaded,
									app.Button().Class("no-padding transparent circle white-text bold").ID(user.Nickname).OnClick(nil).Disabled(c.usersButtonDisabled).Body(
										//app.Text("unshade"),
										app.I().Text("more_horiz"),
									),
								).ElseIf(user.Nickname == c.user.Nickname,
									app.Button().Class("no-padding transparent circle white-text bold").ID(user.Nickname).OnClick(nil).Disabled(true).Body(
										app.I().Text("more_horiz"),
									),
								).Else(
									app.Button().Class("no-padding transparent circle white-text bold").ID(user.Nickname).OnClick(nil).Disabled(c.usersButtonDisabled).Body(
										//app.Text("shade"),
										app.I().Text("more_horiz"),
									),
								),
						),

						// cell's body
						app.Div().Class("row middle-align").Body(
							app.Article().Class("max border thicc").Style("word-break", "break-word").Style("hyphens", "auto").Body(
								app.Span().Text(user.About),
							),
						),

						app.Div().Class("row center-align bottom-padding").Body(

							// If shaded, block any action.
							app.If(shaded, func() app.UI {
								return app.Button().Class("max shrink deep-orange7 white-text bold thicc").Disabled(true).Body(
									app.Text("shaded"),
								)
							}).Else(func() app.UI {
								return app.Div().Body(

									// make button inactive for logged user
									app.If(user.Nickname == c.user.Nickname, func() app.UI {
										return app.Button().Class("max shrink deep-orange7 white-text bold thicc").Disabled(true).Body(
											app.Text("that's you"),
										)
										// if system acc
									}).ElseIf(user.Nickname == "system", func() app.UI {
										return app.Button().Class("max shrink deep-orange7 white-text bold thicc").Disabled(true).Body(
											app.Text("system acc"),
										)
										// private mode
									}).ElseIf(user.Private && !requested && !inFlow, func() app.UI {
										return app.Button().Class("max shrink yellow10 white-text bold thicc").OnClick(c.onClickPrivateOn).Disabled(c.usersButtonDisabled).ID(user.Nickname).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("drafts"),
												app.Text("Ask to follow"),
											),
										)
										// private mode, requested already
									}).ElseIf(user.Private && requested && !inFlow, func() app.UI {
										return app.Button().Class("max shrink grey9 white-text bold thicc").OnClick(c.onClickPrivateOff).Disabled(c.usersButtonDisabled).ID(user.Nickname).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("close"),
												app.Text("Cancel the follow request"),
											),
										)
										// flow toggle off

									}).ElseIf(inFlow, func() app.UI {
										return app.Button().Class("max shrink grey10 white-border white-text bold thicc").ID(user.Nickname).OnClick(c.onClick).Disabled(c.usersButtonDisabled).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("close"),
												app.Text("Unfollow"),
											),
										)
										// flow toggle on
									}).Else(func() app.UI {
										return app.Button().Class("max shrink deep-orange7 white-text bold thicc").ID(user.Nickname).OnClick(c.onClick).Disabled(c.usersButtonDisabled).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("add"),
												app.Text("Follow"),
											),
										)
									}),
								)
							}),

							// shading button
							app.If(shaded, func() app.UI {
								return app.Button().Class("no-padding transparent circular white-text thicc").OnClick(c.onClickUserShade).Disabled(c.userButtonDisabled).ID(user.Nickname).Title("unshade").Body(
									app.I().Text("block"),
								)
							}).ElseIf(user.Nickname == c.user.Nickname, func() app.UI {
								return app.Button().Class("no-padding transparent circular grey white-text thicc").OnClick(nil).Disabled(true).ID(user.Nickname).Title("shading not allowed").Body(
									app.I().Text("block"),
								)
							}).Else(func() app.UI {
								return app.Button().Class("no-padding transparent circular grey white-text thicc").OnClick(c.onClickUserShade).Disabled(c.userButtonDisabled).ID(user.Nickname).Title("shade").Body(
									app.I().Text("block"),
								)
							}),
						),
					),
				)
			}),
		),
	)
}

/*
	return app.Div().Class("post-feed").Body(
		app.Range(p.SortedPolls).Slice(func(idx int) app.UI {
			poll := p.SortedPolls[idx]

			if !p.processPoll(poll) {
				return nil
			}

			return app.Div().Class("post").Body(
				&molecules.PollHeader{
					Poll:                  poll,
					ButtonsDisabled:       p.ButtonsDisabled,
					OnClickLinkActionName: p.OnClickLinkActionName,
				},

				&molecules.PollBody{
					Poll:       poll,
					LoggedUser: p.LoggedUser,
					RenderProps: struct {
						PollTimestamp    string
						UserVoted        bool
						OptionOneShare   int64
						OptionTwoShare   int64
						OptionThreeShare int64
					}{
						PollTimestamp:    p.pollTimestamp,
						UserVoted:        p.userVoted,
						OptionOneShare:   p.optionOneShare,
						OptionTwoShare:   p.optionTwoShare,
						OptionThreeShare: p.optionThreeShare,
					},

					OnClickOptionOneActionName:   p.OnClickOptionOneActionName,
					OnClickOptionTwoActionName:   p.OnClickOptionTwoActionName,
					OnClickOptionThreeActionName: p.OnClickOptionThreeActionName,

					ButtonDisabled:  p.ButtonsDisabled,
					LoaderShowImage: p.LoaderShowImage,
				},

				&molecules.PollFooter{
					Poll:                    poll,
					LoggedUserNickname:      p.LoggedUser.Nickname,
					PollTimestamp:           p.pollTimestamp,
					ButtonsDisabled:         p.ButtonsDisabled,
					OnClickDeleteActionName: p.OnClickDeleteModalShowActionName,
				},
			)
		}),
	)
}*/
