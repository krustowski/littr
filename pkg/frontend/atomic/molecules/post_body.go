package molecules

import (
	"fmt"
	"regexp"

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
		SystemLink      string
	}

	Post models.Post

	OnClickImageActionName   string
	OnClickHistoryActionName string

	ButtonDisabled  bool
	LoaderShowImage bool
}

func (p *PostBody) onClickText(ctx app.Context, e app.Event) {
	if p.Post.ReplyToID == "" {
		return
	}

	ctx.NewActionWithValue(p.OnClickHistoryActionName, p.Post.ReplyToID)
}

func (p *PostBody) Render() app.UI {
	var markupText string

	rgx := regexp.MustCompile(`#[\w]+`)
	markupText = rgx.ReplaceAllStringFunc(p.Post.Content, func(match string) string {
		return fmt.Sprintf(`#link to='/flow/hashtags/%s' class='primary-text'#%s##link#`, match[1:], match)
	})

	rgx = regexp.MustCompile(`https?://[^\s]+`)
	markupText = rgx.ReplaceAllStringFunc(markupText, func(match string) string {
		return fmt.Sprintf(`#link target='_blank' to='%s' class='primary-text'#%s##link#`, match, match)
	})

	rgx = regexp.MustCompile(`@[\w]+`)
	markupText = rgx.ReplaceAllStringFunc(markupText, func(match string) string {
		return fmt.Sprintf(`#link to='/flow/users/%s' class='primary-text'#%s##link#`, match[1:], match)
	})

	return app.Div().Body(
		app.If(p.Post.ReplyToID != "", func() app.UI {
			return app.Article().Class("primary-text border primary-border thicc").Style("max-width", "100%").Body(
				app.Div().Class("row max").Body(
					app.If(p.RenderProps.OriginalSummary != "", func() app.UI {
						return &Details{
							Limit:                 40,
							Text:                  p.RenderProps.OriginalContent,
							SpanID:                p.Post.ReplyToID,
							OnClickSpanActionName: p.OnClickHistoryActionName,
						}
					}).ElseIf(len(p.RenderProps.OriginalContent) > 0, func() app.UI {
						return app.Span().ID(p.Post.ReplyToID).Class("max bold").Text(p.RenderProps.OriginalContent).Style("word-break", "break-word").Style("hyphens", "auto").OnClick(p.onClickText)
					}),

					/*&atoms.Button{
						ID:                p.Post.ReplyToID,
						Title:             "link to original post",
						Class:             "transparent circle",
						Icon:              "history",
						OnClickActionName: p.OnClickHistoryActionName,
						Disabled:          p.ButtonDisabled,
					},*/
				),
			)
		}),

		app.If(p.Post.Nickname == "system", func() app.UI {
			return app.Article().Class("border blue-border bold center-align thicc info").Style("max-width", "100%").Body(
				app.A().Href(p.RenderProps.SystemLink).Body(
					app.Span().Text(p.Post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
				),
			)
		}),

		app.If(len(p.Post.Content) > 0 && p.Post.Nickname != "system", func() app.UI {
			return app.Article().Class("border thicc").Style("max-width", "100%").Body(
				app.If(p.RenderProps.PostSummary != "", func() app.UI {
					return app.Details().Body(
						app.Summary().Text(p.RenderProps.PostSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
						app.Div().Class("space"),
						app.Span().Text(p.Post.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
					)
				}).ElseIf(len(markupText) > 0, func() app.UI {
					return &atoms.Text{
						FormattedText: markupText,
					}
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
					ID:                "img-" + p.Post.ID,
					Src:               p.RenderProps.ImageSource,
					Class:             "no-padding center",
					OnClickActionName: p.OnClickImageActionName,
					Styles:            map[string]string{"max-height": "100%", "max-width": "100%"},
					Attr:              map[string]string{"loading": "lazy"},
				},
			)
		}),
	)
}
