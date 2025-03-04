package organisms

import (
	"fmt"

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

	OnClickUserActionName      string
	OnClickUnfollowActionName  string
	OnClickAskActionName       string
	OnClickCancelActionName    string
	OnClickFollowActionName    string
	OnClickNicknameActionName  string
	OnClickPostCountActionName string
	OnClickShadeActionName     string

	OnMouseEnterActionName string
	OnMouseLeaveActionName string

	// Per user vars. Modified via the processUser() function.
	isInFlow    bool
	isPrivate   bool
	isRequested bool
	isShaded    bool
}

func (u *UserFeed) processUser(user models.User) bool {
	u.isInFlow = false
	u.isPrivate = false
	u.isRequested = false
	u.isShaded = false

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

	if user.Private || user.Options["private"] {
		u.isPrivate = true
	}

	if !user.Searched || user.Nickname == "system" {
		return false
	}

	//log.Printf("user: %s, isInFlow: %t, isShaded: %t, isRequested: %t, isPrivate: %t\n", user.Nickname, u.isInFlow, u.isShaded, u.isRequested, u.isPrivate)

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
							SpanID:                 fmt.Sprintf("%s-span", user.Nickname),
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
							SpanID:                 fmt.Sprintf("%s-span", user.Nickname),
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
						Count:             u.UserStats[user.Nickname].PostCount,
						ID:                user.Nickname,
						Title:             "post count",
						Icon:              "news",
						OnClickActionName: u.OnClickPostCountActionName,
					},
				),

				app.Div().Class("row middle-align").Body(
					app.Article().Class("max border thicc").Style("word-break", "break-word").Style("hyphens", "auto").Body(
						app.Span().Text(user.About),
					),
				),

				app.Div().Class("space"),

				//
				// Follow and shade buttons.
				//

				app.If(user.Nickname == "system", func() app.UI {
					return nil
				}),

				app.If(user.Nickname == u.LoggedUser.Nickname, func() app.UI {
					// When the user is followed.
					return app.Div().Class("row").Body(
						&atoms.Button{
							Class:    "max responsive shrink grey white-text thicc",
							Disabled: true,
							Text:     "That's you",
							Icon:     "cruelty_free",
						},

						&atoms.Button{
							Class:    "max responsive shrink grey10 white-border white-text bold thicc",
							Disabled: true,
							Icon:     "close",
							Text:     "That's you",
						},
					)
				}).ElseIf(u.isShaded, func() app.UI {
					// When the user is shaded, all actions are blocked by default.
					return app.Div().Class("row").Body(
						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey white-text thicc",
							Disabled:          !u.isShaded,
							Text:              "Unshade",
							Icon:              "cruelty_free",
							OnClickActionName: u.OnClickShadeActionName,
						},

						&atoms.Button{
							ID:       user.Nickname,
							Class:    "max responsive shrink grey10 white-border white-text bold thicc",
							Disabled: u.isShaded,
							Icon:     "close",
							Text:     "Shaded",
						},
					)
				}).ElseIf(u.isInFlow, func() app.UI {
					// When the user is followed.
					return app.Div().Class("row").Body(
						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey white-text thicc",
							Disabled:          u.isShaded,
							Text:              "Shade",
							Icon:              "block",
							OnClickActionName: u.OnClickShadeActionName,
						},

						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey10 white-border white-text bold thicc",
							Disabled:          u.ButtonsDisabled,
							Icon:              "close",
							Text:              "Unfollow",
							OnClickActionName: u.OnClickUnfollowActionName,
						},
					)
				}).ElseIf(!u.isInFlow && u.isPrivate && !u.isRequested, func() app.UI {
					return app.Div().Class("row").Body(
						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey white-text thicc",
							Disabled:          u.isShaded,
							Text:              "Shade",
							Icon:              "block",
							OnClickActionName: u.OnClickShadeActionName,
						},

						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink yellow10 white-border white-text bold thicc",
							Disabled:          u.ButtonsDisabled,
							Icon:              "drafts",
							Text:              "Ask to follow",
							OnClickActionName: u.OnClickAskActionName,
						},
					)
				}).ElseIf(!u.isInFlow && u.isRequested && u.isPrivate, func() app.UI {
					return app.Div().Class("row").Body(
						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey white-text thicc",
							Disabled:          u.isShaded,
							Text:              "Shade",
							Icon:              "block",
							OnClickActionName: u.OnClickShadeActionName,
						},

						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey10 white-border white-text bold thicc",
							Disabled:          u.ButtonsDisabled,
							Icon:              "close",
							Text:              "Cancel the follow request",
							OnClickActionName: u.OnClickCancelActionName,
						},
					)
				}).Else(func() app.UI {
					return app.Div().Class("row").Body(
						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink grey white-text thicc",
							Disabled:          u.isShaded,
							Text:              "Shade",
							Icon:              "block",
							OnClickActionName: u.OnClickShadeActionName,
						},

						&atoms.Button{
							ID:                user.Nickname,
							Class:             "max responsive shrink deep-orange7 white-border white-text bold thicc",
							Disabled:          u.ButtonsDisabled,
							Icon:              "add",
							Text:              "Follow",
							OnClickActionName: u.OnClickFollowActionName,
						},
					)
				}),
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
*/
