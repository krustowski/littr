package frontend

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"

	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type StatsPage struct {
	app.Compo
}

type statsContent struct {
	app.Compo

	flowStats map[string]int
	userStats map[string]userStat

	nicknames []string

	users map[string]models.User

	searchString string

	loaderShow bool
	toastShow  bool
	toastText  string
}

type userStat struct {
	// PostCount is a number of posts of such user.
	PostCount int `default:0`

	// ReactionCount tells the number of interactions (stars given).
	ReactionCount int `default:0`

	// FlowerCount is basically a number of followers.
	FlowerCount int `default:0`

	// ShadeCount is basically a number of blockers.
	ShadeCount int `default:0`

	// Searched is a special boolean used by the search engine to mark who is to be shown in search results.
	Searched bool `default:true`
}

func (p *StatsPage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("stats / littr")
}

func (p *StatsPage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&statsContent{},
	)
}

func (c *statsContent) onClickUserFlow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	//c.buttonDisabled = true

	ctx.Navigate("/flow/user/" + key)
}

func (c *statsContent) onSearch(ctx app.Context, e app.Event) {
	val := ctx.JSSrc().Get("value").String()

	//if c.searchString == "" {
	//if val == "" {
	//	return
	//}

	if len(val) > 20 {
		return
	}

	ctx.NewActionWithValue("search", val)
}

func (c *statsContent) handleSearch(ctx app.Context, a app.Action) {
	matchedList := []string{}

	val, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Async(func() {
		users := c.userStats

		// iterate over calculated stats' "rows" and find matchings
		for key, user := range users {
			//user := users[key]
			user.Searched = false

			// use lowecase to search across UNICODE letters
			lval := strings.ToLower(val)
			lkey := strings.ToLower(key)

			if strings.Contains(lkey, lval) {
				user.Searched = true

				//matchedList = append(matchedList, key)
			}

			users[key] = user
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.userStats = users
			c.nicknames = matchedList

			c.loaderShow = false
			log.Println("search dispatch")
		})
		return
	})

}

func (c *statsContent) dismissToast(ctx app.Context, e app.Event) {
	c.toastShow = false
	c.toastText = ""
}

func (c *statsContent) OnMount(ctx app.Context) {
	c.loaderShow = true
	ctx.Handle("search", c.handleSearch)
}

func (c *statsContent) OnNav(ctx app.Context) {
	toastText := ""

	ctx.Async(func() {
		payload := struct {
			FlowStats map[string]int         `json:"flow_stats"`
			UserStats map[string]userStat    `json:"user_stats"`
			Users     map[string]models.User `json:"users"`
			Code      int                    `json:"code"`
		}{}

		// fetch the stats
		if byteData, _ := littrAPI("GET", "/api/v1/stats", nil, "", 0); byteData != nil {
			err := json.Unmarshal(*byteData, &payload)
			if err != nil {
				log.Println(err.Error())

				ctx.Dispatch(func(ctx app.Context) {
					c.toastText = "backend error: " + err.Error()
					c.toastShow = (c.toastText != "")
				})
				return
			}
		} else {
			log.Println("cannot fetch stats")

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = "cannot fetch stats"
				c.toastShow = (c.toastText != "")
			})
			return
		}

		if payload.Code == 401 {
			toastText = "please log-in again"

			ctx.LocalStorage().Set("user", "")
			ctx.LocalStorage().Set("authGranted", false)

			ctx.Dispatch(func(ctx app.Context) {
				c.toastText = toastText
				c.toastShow = (toastText != "")
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			c.users = payload.Users
			c.flowStats = payload.FlowStats
			c.userStats = payload.UserStats

			c.loaderShow = false
			log.Println("dispatch ends")
		})
		return
	})
}

func (c *statsContent) Render() app.UI {
	users := c.userStats
	flowStats := c.flowStats

	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("user stats"),
				//app.P().Text("wanna know your flow stats? how many you got in the flow and vice versa? yo"),
			),
		),
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

		app.Div().Class("field prefix round fill").Style("border-radius", "8px").Body(
			app.I().Class("front").Text("search"),
			//app.Input().Type("search").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),
			//app.Input().ID("search").Type("text").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),
			app.Input().ID("search").Type("text").OnChange(c.onSearch).OnSearch(c.onSearch),
		),

		app.Table().Class("border right-align").ID("table-stats-flow").Body(
			// table header
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("left-align").Body(
						app.Span().Style("writing-mode", "vertical-lr").Text("nickname"),
					),
					app.Th().Class("right-align no-padding").Body(
						app.Span().Style("writing-mode", "vertical-lr").Text("posts"),
					),
					app.Th().Class("right-align no-padding").Body(
						app.Span().Style("writing-mode", "vertical-lr").Text("stars"),
					),
					app.Th().Class("right-align no-padding").Body(
						app.Span().Style("writing-mode", "vertical-lr").Text("flowers"),
					),
					app.Th().Class("right-align no-padding").Body(
						app.Span().Style("writing-mode", "vertical-lr").Text("shades"),
					),
					app.Th().Class("right-align no-padding").Body(
						app.Span().Style("writing-mode", "vertical-lr").Text("ratio"),
					),
				),
			),

			// table body
			app.TBody().Body(
				app.Range(users).Map(func(key string) app.UI {
					// calculate the ratio
					ratio := func() float64 {
						if users[key].PostCount <= 0 {
							return 0
						}

						stars := float64(users[key].ReactionCount)
						posts := float64(users[key].PostCount)
						shades := float64(users[key].ShadeCount)
						users := float64(flowStats["users"])

						baseRatio := stars / posts
						shadeCoeff := 1.0

						if users > 1 && shades > 1 {
							shadeCoeff = 1 - math.Log(shades)/math.Log(users)
						}

						return baseRatio * shadeCoeff
					}()

					// filter out unmatched keys
					//log.Printf("%s: %t\n", key, users[key].Searched)

					if !users[key].Searched || c.users[key].Nickname == "system" {
						return app.P().Text("")
					}

					return app.Tr().Body(
						app.Td().Class("left-align").Body(
							app.P().Body(
								app.P().Body(
									//app.B().Text(key).Class("deep-orange-text"),
									app.A().Class("bold deep-orange-text").OnClick(c.onClickUserFlow).Text(key).ID(key),
									//app.A().Class("bold deep-orange-text").OnClick(nil).Text(key).ID(key),
								),
							),
						),
						app.Td().Class("right-align").Body(
							app.P().Body(
								app.Text(strconv.FormatInt(int64(users[key].PostCount), 10)),
							),
						),
						app.Td().Class("right-align").Body(
							app.P().Body(
								app.Text(strconv.FormatInt(int64(users[key].ReactionCount), 10)),
							),
						),
						app.Td().Class("right-align").Body(
							app.P().Body(
								app.Text(strconv.FormatInt(int64(users[key].FlowerCount), 10)),
							),
						),
						app.Td().Class("right-align").Body(
							app.P().Body(
								app.Text(strconv.FormatInt(int64(users[key].ShadeCount), 10)),
							),
						),
						app.Td().Class("right-align").Body(
							app.P().Body(
								app.Text(strconv.FormatFloat(ratio, 'f', 2, 64)),
							),
						),
					)
				}),
			),
		),
		app.If(c.loaderShow,
			app.Div().Class("small-space"),
			app.Progress().Class("circle center large deep-orange-border active"),
		),

		app.Div().Class("large-space"),

		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("system stats"),
				//app.P().Text("pop in to see how much this instance lit nocap"),
			),
		),
		//app.P().Body(
		//),
		app.Div().Class("space"),

		app.Table().Class("border left-align").ID("table-stats-system").Body(
			// table header
			app.THead().Body(
				app.Tr().Body(
					app.Th().Class("left align").Text("property"),
					app.Th().Class("right-align").Text("value"),
				),
			),
			// table body
			app.TBody().Body(
				app.Range(flowStats).Map(func(key string) app.UI {
					return app.Tr().Body(
						app.Td().Class("left-align").Body(
							app.P().Body(
								app.P().Body(
									app.B().Text(key).Class("deep-orange-text"),
								),
							),
						),
						app.Td().Class("right-align").Body(
							app.P().Body(
								app.Text(strconv.FormatInt(int64(flowStats[key]), 10)),
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
