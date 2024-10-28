package welcome

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Article().Body(
			app.Div().Class("").Body(
				app.Div().Class("row center-align").Body(
					app.Img().Src("/web/android-chrome-192x192.png"),
					app.H4().Body(
						app.Span().Body(
							app.Text("littr"),
						),
					),
				),
				app.Div().Class("space"),
				app.H6().Class("margin-bottom center-align").Body(
					app.Span().Body(
						app.Text("a simple nanoblogging platform"),
					),
				),
			),

			app.Div().Class("row no-margin  large-padding").Body(
				//app.I().Text("lightbulb").Class("amber-text"),
				app.P().Class("max").Body(
					app.Span().Class("deep-orange-text").Text("welcome to "),
					app.Span().Class("deep-orange-text bold").Text("littr"),
					app.Span().Class("deep-orange-text").Text("! "),
					app.Span().Text("this site acts as a simple platform for anyone who likes to post short notes, messages, daydreaming ideas and more! you can use it as a personal journal charting your journey through life that can be shared with other accounts"),
					app.Div().Class("small-space"),

					app.Span().Text("the very main page of this platform is called"),
					app.Span().Class("deep-orange-text bold").Text(" flow "),
					app.Span().Text("(shown below); this page lists all your posts in reverse chronological order (newest to oldest) plus posts from other folks/accounts that you have added to your flow"),
					app.Div().Class("small-space"),

					app.Span().Text("to navigate to the "),
					app.Span().Class("bold").Text("login "),
					app.Span().Text("page (where the link to "),
					app.Span().Class("bold").Text("registration "),
					app.Span().Text("sits as well) use the "),
					app.I().Class("small").Class("deep-orange-text").Body(
						app.Text("login"),
					),
					app.Span().Text(" button in the upper right corner"),
				),
			),
		),

		app.Article().Class("center-align center").Body(
			app.H6().Class("margin-bottom center-align").Body(
				app.Span().Body(
					app.Text("flow page"),
				),
			),
			app.Div().Class("space"),

			app.Div().Style("z-index", "5").Class("medium no-padding center-align").Body(
				app.Img().Class("center-align bottom lazy").Src("https://krusty.space/littr_flow_new_post_live_v0.30.17.jpg").Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
			),
		),

		app.Div().Class("medium-space"),
	)
}
