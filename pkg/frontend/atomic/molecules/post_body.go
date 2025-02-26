package molecules

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type PostBody struct {
	app.Compo

	RenderProps struct {
		ImageSource     string
		PostSummary     string
		OriginalContent string
		OriginalSummary string
		PostTimestamp   string
	}

	Post models.Post

	OnClickImage app.EventHandler
	OnClickLink  app.EventHandler

	ButtonDisabled  bool
	LoaderShowImage bool
}

func (p *PostBody) OnMount(ctx app.Context) {}

func (p *PostBody) Render() app.UI {
	return app.Div().Body(
		app.If(p.Post.ReplyToID != "", func() app.UI {
			return app.Article().Class("black-text border reply thicc").Style("max-width", "100%").Body(
				app.Div().Class("row max").Body(
					app.If(p.RenderProps.OriginalSummary != "", func() app.UI {
						return app.Details().Class("max").Body(
							app.Summary().Text(p.RenderProps.OriginalSummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
							app.Div().Class("space"),
							app.Span().Class("bold").Text(p.RenderProps.OriginalContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
						)
					}).ElseIf(len(p.RenderProps.OriginalContent) > 0, func() app.UI {
						return app.Span().Class("max bold").Text(p.RenderProps.OriginalContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line")
					}),

					&atoms.Button{
						ID:       p.Post.ReplyToID,
						Title:    "link to original post",
						Class:    "transparent circle",
						Icon:     "history",
						OnClick:  p.OnClickLink,
						Disabled: p.ButtonDisabled,
					},
				),
			)
		}),

		app.If(len(p.Post.Content) > 0, func() app.UI {
			return app.Article().Class("border thicc").Style("max-width", "100%").Body(
				app.If(p.RenderProps.PostSummary != "", func() app.UI {
					return app.Details().Body(
						app.Summary().Text(p.RenderProps.PostSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
						app.Div().Class("space"),
						app.Span().Text(p.Post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
					)
				}).Else(func() app.UI {
					return app.Span().Text(p.Post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line")
				}),
			)
		}),

		app.If(p.Post.Figure != "" && p.Post.Nickname != "system", func() app.UI {
			return app.Article().Style("z-index", "4").Class("transparent medium thicc").Body(
				&atoms.Loader{
					ShowLoader: p.LoaderShowImage,
				},

				&atoms.Image{
					ID:      p.Post.ID,
					Src:     p.RenderProps.ImageSource,
					Class:   "no-padding center middle lazy",
					OnClick: p.OnClickImage,
					Styles:  map[string]string{"max-height": "100%", "max-width": "100%"},
					Attr:    map[string]string{"loading": "lazy"},
				},
			)
		}),
	)
}
