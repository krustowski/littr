package stats

import (
	"math"
	"strconv"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) Render() app.UI {
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
		app.A().Href(c.toast.TLink).OnClick(c.onDismissToast).Body(
			app.If(c.toast.TText != "", func() app.UI {
				return app.Div().Class("snackbar white-text top active "+common.ToastColor(c.toast.TType)).Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				)
			}),
		),

		app.Div().Class("field prefix round fill").Style("border-radius", "8px").Body(
			app.I().Class("front").Text("search"),
			//app.Input().Type("search").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),
			//app.Input().ID("search").Type("text").OnChange(c.ValueTo(&c.searchString)).OnSearch(c.onSearch),
			app.Input().ID("search").Type("text").OnChange(c.onSearch).OnSearch(c.onSearch),
		),

		app.Table().Class("border sortable right-align").ID("table-stats-flow").Body(
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
									app.A().Class("bold primary-text").OnClick(c.onClickUserFlow).Text(key).ID(key),
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
		app.If(c.loaderShow, func() app.UI {
			return app.Div().Body(
				app.Div().Class("small-space"),
				app.Progress().Class("circle center large primary-border active"),
			)
		}),

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
									app.B().Text(key).Class("primary-text"),
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
		app.If(c.loaderShow, func() app.UI {
			return app.Div().Body(
				app.Div().Class("small-space"),
				app.Progress().Class("circle center large primary-border active"),
			)
		}),
	)
}
