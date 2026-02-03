package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
)

type ModalAppInfo struct {
	app.Compo

	ShowModal bool

	SseConnectionStatus string

	OnClickDismissActionName string
	OnClickReloadActionName  string
}

func (m *ModalAppInfo) Render() app.UI {
	return app.Div().Body(
		app.If(m.ShowModal, func() app.UI {
			return app.Dialog().ID("info-modal").Class("grey10 white-text center-align active thicc center").Body(
				app.Article().Class("row white-text center-align border thicc").Body(

					&atoms.Image{
						Styles: map[string]string{"max-width": "10em"},
						Src:    "/web/android-chrome-512x512.png",
					},

					app.H4().Body(
						app.Span().Body(
							app.Text("littr"),
							app.If(app.Getenv("APP_ENVIRONMENT") != "prod", func() app.UI {
								return app.Span().Class("col").Body(
									app.Sup().Body(
										app.If(app.Getenv("APP_ENVIRONMENT") == "stage", func() app.UI {
											return app.Text(" (stage) ")
										}).Else(func() app.UI {
											return app.Text(" (dev) ")
										}),
									),
								)
							}),
						),
					),
				),

				app.Article().Class("center-align large-text border thicc").Body(
					app.P().Body(
						app.A().Class("primary-text bold").Href("/tos").Text("Terms of Service"),
					),
					app.P().Body(
						app.A().Class("primary-text bold").Href("https://krusty.space/projects/littr").Text("Documentation (external)"),
					),
				),

				app.Article().Class("center-align white-text border thicc").Body(
					app.Text("Version: "),
					app.A().Text(app.Getenv("APP_VERSION")).Href("https://github.com/krustowski/littr").Style("font-weight", "bolder"),
					app.P().Body(
						app.Text("SSE status: "),
						app.If(m.SseConnectionStatus == "connected", func() app.UI {
							return app.Span().ID("heartbeat-info-text").Text(m.SseConnectionStatus).Class("green-text bold")
						}).Else(func() app.UI {
							return app.Span().ID("heartbeat-info-text").Text(m.SseConnectionStatus).Class("amber-text bold")
						}),
					),
				),

				app.Nav().Class("center-align").Body(
					app.P().Body(
						app.Text("Powered by "),
						app.A().Href("https://go-app.dev/").Text("go-app").Style("font-weight", "bolder"),
						app.Text(" & "),
						app.A().Href("https://www.beercss.com/").Text("beercss").Style("font-weight", "bolder"),
					),
				),

				app.Div().Class("row").Body(
					&atoms.Button{
						Class:             "max bold black white-text thicc",
						Icon:              "close",
						Text:              "Close",
						OnClickActionName: m.OnClickDismissActionName,
					},

					&atoms.Button{
						Class:             "max bold primary-container white-text thicc",
						Icon:              "refresh",
						Text:              "Reload",
						OnClickActionName: m.OnClickReloadActionName,
					},
				),
			)
		}),
	)
}
