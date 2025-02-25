package organisms

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/models"
)

type ModalPostReply struct {
	app.Compo

	replyPostContent string
	newFigLink       string
	newFigFile       string

	PostOriginal models.Post

	ModalButtonsDisabled bool
	ModalShow            bool

	OnClickDismiss app.EventHandler
	OnClickReply   app.EventHandler
	OnBlur         app.EventHandler
	OnFigureUpload app.EventHandler
}

func (m *ModalPostReply) OnMount(ctx app.Context) {
}

func (m *ModalPostReply) Render() app.UI {
	// compose a summary of a long post to be replied to
	replySummary := ""
	if m.ModalShow && len(m.PostOriginal.Content) > config.MaxPostLength {
		replySummary = m.PostOriginal.Content[:config.MaxPostLength/10] + "- [...]"
	}

	return app.Div().Body(
		app.If(m.ModalShow, func() app.UI {
			return app.Dialog().ID("reply-modal").Class("grey10 white-text center-align active thicc").Style("max-width", "90%").Style("z-index", "75").Body(
				app.Nav().Class("center-align").Body(
					app.H5().Text("reply"),
				),
				app.Div().Class("space"),

				// Original content (text).
				app.If(m.PostOriginal.Content != "", func() app.UI {
					return app.Article().Class("reply black-text border thicc").Style("max-width", "100%").Body(
						app.If(replySummary != "", func() app.UI {
							return app.Details().Body(
								app.Summary().Text(replySummary).Style("word-break", "break-word").Style("hyphens", "auto").Class("italic"),
								app.Div().Class("space"),

								app.Span().Class("bold").Text(m.PostOriginal.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic"),
							)
						}).Else(func() app.UI {
							return app.Span().Class("bold").Text(m.PostOriginal.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic")
						}),
					)
				}),

				app.Div().Class("field label textarea border extra deep-orange-text").Style("border-radius", "8px").Body(
					//app.Textarea().Class("active").Name("replyPost").OnChange(c.ValueTo(&c.replyPostContent)).AutoFocus(true).Placeholder("reply to: "+c.posts[c.interactedPostKey].Nickname),
					app.Textarea().Class("active").Name("replyPost").Text(m.replyPostContent).OnChange(m.ValueTo(&m.replyPostContent)).AutoFocus(true).ID("reply-textarea").OnBlur(m.OnBlur),
					app.Label().Text("Reply to: "+m.PostOriginal.Nickname).Class("active deep-orange-text"),
					//app.Label().Text("text").Class("active"),
				),
				app.Div().Class("field label border extra deep-orange-text").Style("border-radius", "8px").Body(
					app.Input().ID("fig-upload").Class("active").Type("file").OnChange(m.ValueTo(&m.newFigLink)).OnInput(m.OnFigureUpload).Accept("image/*"),
					app.Input().Class("active").Type("text").Value(m.newFigFile).Disabled(true),
					app.Label().Text("Image").Class("active deep-orange-text"),
					app.I().Text("image"),
				),

				// Reply buttons.
				app.Div().Class("row").Body(
					&atoms.Button{
						Class:    "max bold black white-text thicc",
						Icon:     "close",
						Text:     "Cancel",
						OnClick:  m.OnClickDismiss,
						Disabled: m.ModalButtonsDisabled,
					},

					&atoms.Button{
						ID:       "reply",
						Class:    "max bold deep-orange7 white-text thicc",
						Icon:     "reply",
						Text:     "Reply",
						OnClick:  m.OnClickReply,
						Disabled: m.ModalButtonsDisabled,
					},
				),
				app.Div().Class("space"),
			)
		}),
	)
}
