package users

import (
	"sort"
	"time"

	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	keys := []string{}

	// prepare the keys array
	for key := range c.users {
		keys = append(keys, key)
	}

	// sort them keys
	sort.Strings(keys)

	// prepare the sorted users array
	sortedUsers := func() []models.User {
		var sorted []models.User

		for _, key := range keys {
			if !c.users[key].Searched {
				continue
			}

			sorted = append(sorted, c.users[key])
		}

		return sorted
	}()

	var userInModalInfo map[string]string = nil

	if c.showUserPreviewModal {
		userInModalInfo = map[string]string{
			"full name": c.userInModal.FullName,
			"web":       c.userInModal.Web,
			//"e-mail":    c.userInModal.Email,
			"last active": c.userInModal.LastActiveTime.Format("Jan 02, 2006; 15:04:05 -0700"),
			"registered":  c.userInModal.RegisteredTime.Format("Jan 02, 2006; 15:04:05 -0700"),
		}

		//userGravatarURL := getGravatar(c.userInModal.Email)
		//userGravatarURL = c.getGravatarURL()
	}

	// prepare posts according to the actual pagination and pageNo
	pagedUsers := []models.User{}
	//pagedUsers := sortedUsers

	end := len(sortedUsers)
	start := 0

	stop := func(c *Content) int {
		var pos int

		if c.pagination > 0 {
			// (c.pageNo - 1) * c.pagination + c.pagination
			pos = c.pageNo * c.pagination
		}

		if pos > end {
			// kill the scrollEventListener (observers scrolling)
			c.scrollEventListener()
			c.paginationEnd = true

			return (end)
		}

		if pos < 0 {
			return 0
		}

		return pos
	}(c)

	if end > 0 && stop > 0 {
		pagedUsers = sortedUsers[start:stop]
	}

	var numOfReqs int64 = 0

	requestList := c.user.RequestList
	for _, state := range requestList {
		if state {
			numOfReqs++
			// we don't need to loop further as the number is going to be always greater than zero henceforth
			break
		}
	}

	var userRegisteredTime string
	var userLastActiveTime string

	if c.userInModal.Nickname != "" {
		registeredTime := c.userInModal.RegisteredTime
		lastActiveTime := c.userInModal.LastActiveTime

		registered := app.Window().
			Get("Date").
			New(registeredTime.Format(time.RFC3339))

		lastActive := app.Window().
			Get("Date").
			New(lastActiveTime.Format(time.RFC3339))

		userRegisteredTime = registered.Call("toLocaleString", "en-GB").String()
		userLastActiveTime = lastActive.Call("toLocaleString", "en-GB").String()
	}

	return app.Main().Class("responsive").Body(
		app.If(c.user.RequestList != nil && numOfReqs > 0,
			app.Div().Class("row").Body(
				app.Div().Class("max padding").Body(
					app.H5().Text("requests"),
				),
			),
			app.Div().Class("space"),

			// requests table
			app.Table().Class("border").ID("table-users").Style("width", "100%").Body(
				app.TBody().Body(
					app.Range(c.user.RequestList).Map(func(key string) app.UI {
						if !c.user.RequestList[key] {
							return nil
						}

						return app.Tr().Body(
							app.Td().Class("left-align").Body(

								// cell's header
								app.Div().Class("row medium top-padding").Body(
									app.Img().Class("responsive max left").Src(c.users[key].AvatarURL).Style("max-width", "60px").Style("border-radius", "50%"),
									app.P().ID(c.users[key].Nickname).Text(c.users[key].Nickname).Class("deep-orange-text bold max").OnClick(c.onClickUser),
									app.Button().Class("max responsive no-padding transparent circular black white-text thicc").OnClick(c.onClickCancel).Disabled(c.userButtonDisabled).ID(c.users[key].Nickname).Body(
										app.Span().Body(
											app.I().Style("padding-right", "5px").Text("close"),
											app.Text("Cancel"),
										),
									),
									app.Button().Class("max responsive no-padding transparent circular deep-orange7 white-text thicc").OnClick(c.onClickAllow).Disabled(c.userButtonDisabled).ID(c.users[key].Nickname).Body(
										app.Span().Body(
											app.I().Style("padding-right", "5px").Text("check"),
											app.Text("Allow"),
										),
									),
								),
							),
						)
					}),
				),
			),
			app.Div().Class("space"),
		),

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("flowers"),
			),
		),
		app.Div().Class("space"),

		// user info modal
		app.If(c.showUserPreviewModal && userInModalInfo != nil,
			app.Dialog().ID("user-modal").Class("grey9 white-text center-align active border thicc").Style("max-width", "90%").Body(

				//app.Img().Class("small-width small-height").Src(c.userInModal.AvatarURL),
				app.Img().Class("small-width").Src(c.userInModal.AvatarURL).Style("max-width", "120px").Style("border-radius", "50%"),

				app.Div().Class("row center-align").Body(
					app.H5().Class().Body(
						app.A().Href("/flow/users/"+c.userInModal.Nickname).Text(c.userInModal.Nickname),
					),

					app.If(c.userInModal.Web != "",
						app.A().Href(c.userInModal.Web).Body(
							app.Span().Class("bold").Body(
								app.I().Text("captive_portal"),
							),
						),
					),
				),

				app.If(c.userInModal.About != "",
					app.Article().Class("center-align border thicc").Style("word-break", "break-word").Style("hyphens", "auto").Text(c.userInModal.About),
				),

				app.Article().Class("left-align border thicc").Body(
					app.P().Class("bold").Text("Registered"),
					app.P().Class().Text(userRegisteredTime),

					app.P().Class("bold").Text("Last online"),
					app.P().Class().Text(userLastActiveTime),
				),

				//app.Div().Class("large-space"),
				app.Div().Class("row center-align").Body(
					app.Button().Class("max black white-text thicc").OnClick(c.onDismissToast).Body(
						app.Span().Body(
							app.I().Style("padding-right", "5px").Text("close"),
							app.Text("Close"),
						),
					),
				),
			),
		),

		// search bar
		app.Div().Class("field prefix round fill thicc").Body(
			app.I().Class("front").Text("search"),
			//app.Input().Type("search").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),
			app.Input().ID("search").Type("text").OnChange(c.onSearch).OnSearch(c.onSearch),
		),

		// users table
		app.Table().Class("border").ID("table-users").Style("width", "100%").Style("border-spacing", "0.1em").Style("padding", "0 0 2em 0").Body(
			app.TBody().Body(
				app.Range(pagedUsers).Slice(func(idx int) app.UI {
					//app.Range(c.users).Map(func(idx string) app.UI {
					user := pagedUsers[idx]
					//user := c.users[idx]

					var inFlow bool = false
					var shaded bool = false
					var requested bool = false
					var found bool

					if c.user.FlowList != nil {
						if inFlow, found = c.user.FlowList[user.Nickname]; found && inFlow {
							inFlow = true
						}
					}

					if c.user.ShadeList != nil {
						if shaded, found = c.user.ShadeList[user.Nickname]; found && shaded {
							shaded = true
						}
					}

					if user.RequestList != nil {
						if requested, found = user.RequestList[c.user.Nickname]; !found {
							requested = false
						}
					}

					if !user.Searched || user.Nickname == "system" {
						return nil
					}

					return app.Tr().Body(
						app.Td().Class("left-align").Body(

							// cell's header
							app.Div().Class("row medium top-padding").Body(
								app.Img().Class("responsive max left").Src(user.AvatarURL).Style("max-width", "60px").Style("border-radius", "50%"),

								app.If(user.Private,
									// nasty hack to ensure the padding lock icon is next to nickname
									app.P().ID(user.Nickname).Class("deep-orange-text bold").OnClick(c.onClickUser).Body(
										app.Span().Class("large-text bold deep-orange-text").Text(user.Nickname),
									),

									// show private mode
									app.Span().Class("bold max").Body(
										app.I().Text("lock"),
									),
								).Else(
									app.P().ID(user.Nickname).Class("deep-orange-text bold max").OnClick(c.onClickUser).Body(
										app.Span().Class("large-text bold deep-orange-text").Text(user.Nickname),
									),
								),

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
								*/
							),

							// cell's body
							app.Div().Class("row middle-align").Body(

								app.Article().Class("max border thicc").Style("word-break", "break-word").Style("hyphens", "auto").Body(
									app.Span().Text(user.About),
								),
							),

							app.Div().Class("row center-align bottom-padding").Body(

								// If shaded, block any action.
								app.If(shaded,
									app.Button().Class("max shrink deep-orange7 white-text bold thicc").Disabled(true).Body(
										app.Text("shaded"),
									),
								).Else(

									// make button inactive for logged user
									app.If(user.Nickname == c.user.Nickname,
										app.Button().Class("max shrink deep-orange7 white-text bold thicc").Disabled(true).Body(
											app.Text("that's you"),
										),
									// if system acc
									).ElseIf(user.Nickname == "system",
										app.Button().Class("max shrink deep-orange7 white-text bold thicc").Disabled(true).Body(
											app.Text("system acc"),
										),
									// private mode
									).ElseIf(user.Private && !requested && !inFlow,
										app.Button().Class("max shrink yellow10 white-text bold thicc").OnClick(c.onClickPrivateOn).Disabled(c.usersButtonDisabled).ID(user.Nickname).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("drafts"),
												app.Text("Ask to follow"),
											),
										),
									// private mode, requested already
									).ElseIf(user.Private && requested && !inFlow,
										app.Button().Class("max shrink grey9 white-text bold thicc").OnClick(c.onClickPrivateOff).Disabled(c.usersButtonDisabled).ID(user.Nickname).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("close"),
												app.Text("Cancel the follow request"),
											),
										),
									// flow toggle off
									).ElseIf(inFlow,
										app.Button().Class("max shrink border black white-border white-text bold thicc").ID(user.Nickname).OnClick(c.onClick).Disabled(c.usersButtonDisabled).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("close"),
												app.Text("Unfollow"),
											),
										),
									// flow toggle on
									).Else(
										app.Button().Class("max shrink deep-orange7 white-text bold thicc").ID(user.Nickname).OnClick(c.onClick).Disabled(c.usersButtonDisabled).Body(
											app.Span().Body(
												app.I().Style("padding-right", "5px").Text("add"),
												app.Text("Follow"),
											),
										),
									),
								),

								// shading button
								app.If(shaded,
									app.Button().Class("no-padding transparent circular white-text thicc").OnClick(c.onClickUserShade).Disabled(c.userButtonDisabled).ID(user.Nickname).Title("unshade").Body(
										app.I().Text("block"),
									),
								).ElseIf(user.Nickname == c.user.Nickname,
									app.Button().Class("no-padding transparent circular grey white-text thicc").OnClick(nil).Disabled(true).ID(user.Nickname).Title("shading not allowed").Body(
										app.I().Text("block"),
									),
								).Else(
									app.Button().Class("no-padding transparent circular grey white-text thicc").OnClick(c.onClickUserShade).Disabled(c.userButtonDisabled).ID(user.Nickname).Title("shade").Body(
										app.I().Text("block"),
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
