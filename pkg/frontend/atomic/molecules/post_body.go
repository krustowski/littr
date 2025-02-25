package molecules

import (
	"net/url"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type PostBody struct {
	app.Compo

	imgSrc                 string
	hideReplies            bool
	postDetailsSummary     string
	previousDetailsSummary string
	previousContent        string

	Post         models.Post
	PostOriginal models.Post
	//Posts        map[string]models.Post

	OnClickImage app.EventHandler
	OnClickLink  app.EventHandler

	ButtonDisabled  bool
	LoaderShowImage bool

	ReplyToID string
}

func (p *PostBody) OnMount(ctx app.Context) {
	if len(p.Post.Content) > config.MaxPostLength {
		p.postDetailsSummary = p.Post.Content[:config.MaxPostLength/10] + "- [...]"
	}

	if p.Post.ReplyToID != "" {
		if !p.hideReplies {
			/*if _, found := p.Posts[p.Post.ReplyToID]; found {
			}*/
		}
	}

	// Fetch the image source
	// Check the URL/URI format
	switch p.Post.Type {
	case "fig":
		if _, err := url.ParseRequestURI(p.Post.Content); err == nil {
			p.imgSrc = p.Post.Content
		} else {
			fileExplode := strings.Split(p.Post.Content, ".")
			extension := fileExplode[len(fileExplode)-1]

			p.imgSrc = "/web/pix/thumb_" + p.Post.Content
			if extension == "gif" {
				p.imgSrc = "/web/click-to-see-gif.jpg"
			}
		}
	case "post":
		if _, err := url.ParseRequestURI(p.Post.Figure); err == nil {
			p.imgSrc = p.Post.Figure
		} else {
			fileExplode := strings.Split(p.Post.Figure, ".")
			extension := fileExplode[len(fileExplode)-1]

			p.imgSrc = "/web/pix/thumb_" + p.Post.Figure
			if extension == "gif" {
				p.imgSrc = "/web/click-to-see.gif"
			}
		}
	}
}

func (p *PostBody) Render() app.UI {
	return app.Div().Body(
		app.If(p.Post.ReplyToID != "", func() app.UI {
			return app.Article().Class("black-text border reply thicc").Style("max-width", "100%").Body(
				app.Div().Class("row max").Body(
					app.If(p.previousDetailsSummary != "", func() app.UI {
						return app.Details().Class("max").Body(
							app.Summary().Text(p.previousDetailsSummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
							app.Div().Class("space"),
							app.Span().Class("bold").Text(p.previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line"),
						)
					}).Else(func() app.UI {
						return app.Span().Class("max bold").Text(p.previousContent).Style("word-break", "break-word").Style("hyphens", "auto").Style("white-space", "pre-line")
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
				app.If(p.postDetailsSummary != "", func() app.UI {
					return app.Details().Body(
						app.Summary().Text(p.postDetailsSummary).Style("hyphens", "auto").Style("word-break", "break-word"),
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
				app.If(p.LoaderShowImage, func() app.UI {
					return app.Div().Body(
						app.Div().Class("small-space"),
						app.Div().Class("loader center large deep-orange active"),
					)
				}),

				app.Img().Class("no-padding center middle lazy").Src(p.imgSrc).Style("max-width", "100%").Style("max-height", "100%").Attr("loading", "lazy").OnClick(p.OnClickImage).ID(p.Post.ID),
				/*&atoms.Image{
					ID:    p.Post.ID,
					Src:   p.imgSrc,
					Class: "no-padding center middle lazy",
					//Width:   "100%",
					//Height:  "100%",
					Radius:  "100%",
					OnClick: p.OnClickImage,
					Attr:    map[string]string{"loading": "lazy"},
				},*/
			)
		}),
	)
}
