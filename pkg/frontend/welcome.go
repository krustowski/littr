package frontend

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type WelcomePage struct {
	app.Compo

	mode string
}

type welcomeContent struct {
	app.Compo
}

func (p *WelcomePage) Render() app.UI {
	return app.Div().Body(
		&header{},
		&footer{},
		&welcomeContent{},
	)
}

func (p *WelcomePage) OnNav(ctx app.Context) {
	ctx.Page().SetTitle("welcome / littr")
}

func (c *welcomeContent) OnMount(ctx app.Context) {
}

func (c *welcomeContent) OnNav(ctx app.Context) {
}

func (c *welcomeContent) Render() app.UI {
	return app.Main().Class("responsive").Body(
		app.Article().Class("").Style("border-radius", "8px").Body(
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
		app.Div().Class("space"),

		app.Article().Class("row large-padding").Body(
			app.I().Text("lightbulb").Class("amber-text"),
			app.P().Class("max").Body(
				app.Span().Class("deep-orange-text").Text("welcome to littr! "),
				app.Span().Text("this site acts as a simple platform for anyone who likes to post short notes, messages, daydreaming ideas and more! you can use it as a personal journal charting your journey through life that can be shared with other accounts"),
				app.Div().Class("small-space"),

				app.Span().Text("the very main page of this platform is called"),
				app.Span().Class("deep-orange-text").Text(" flow "),
				app.Span().Text("(shown below); this page lists all your posts in reverse chronological order (newest to oldest) plus posts from other folks/accounts that you have added to your flow"),
				app.Div().Class("small-space"),

				app.Span().Text("to navigate to the login page (where the link to registration sits as well) use the icon/button in the upper right corner: "),
				app.I().Class("large").Class("deep-orange-text").Body(
					app.Text("login"),
				),
			),
		),

		app.Article().Style("z-index", "5").Style("border-radius", "8px").Class("medium no-padding transparent center-align").Body(
			app.Img().Class("absolute margin-top center middle lazy").Src("https://krusty.space/littr_flow_new_post_live_v0.30.17.jpg").Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy"),
		),

		app.Div().Class("medium-space"),
	)
}
