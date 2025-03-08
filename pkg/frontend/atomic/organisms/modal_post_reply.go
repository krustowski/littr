package organisms

import (
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/config"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
	"go.vxn.dev/littr/pkg/models"
)

type ModalPostReply struct {
	app.Compo

	ReplyPostContent *string
	ImageData        *[]byte
	ImageLink        *string
	ImageFile        *string

	PostOriginal models.Post

	ModalButtonsDisabled *bool
	ModalShow            bool

	OnClickDismissActionName string
	OnClickReplyActionName   string
	OnBlurActionName         string
	OnFigureUploadActionName string
	//OnFigureUpload app.EventHandler
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
					return app.Article().Class("primary-text border primary-border thicc").Style("max-width", "100%").Body(
						app.If(replySummary != "", func() app.UI {
							return &molecules.Details{
								Text:  m.PostOriginal.Content,
								Limit: 40,
							}
						}).Else(func() app.UI {
							return app.Span().Class("bold").Text(m.PostOriginal.Content).Style("word-break", "break-word").Style("hyphens", "auto").Style("font-type", "italic")
						}),
					)
				}),

				&atoms.Textarea{
					ID:               "reply-textarea",
					Class:            "field label textarea border extra blue-text thicc",
					ContentPointer:   m.ReplyPostContent,
					Name:             "replyPost",
					LabelText:        fmt.Sprintf("Reply to: %s", m.PostOriginal.Nickname),
					OnBlurActionName: "blur",
				},

				&molecules.ImageInput{
					ImageData:            m.ImageData,
					ImageFile:            m.ImageFile,
					ImageLink:            m.ImageLink,
					ButtonsDisabled:      m.ModalButtonsDisabled,
					LocalStorageFileName: "newReplyImageFile",
					LocalStorageDataName: "newReplyImageData",
				},

				// Reply buttons.
				app.Div().Class("row").Body(
					&atoms.Button{
						Class:             "max bold black white-text thicc",
						Icon:              "close",
						Text:              "Cancel",
						OnClickActionName: m.OnClickDismissActionName,
						Disabled:          *m.ModalButtonsDisabled,
					},

					&atoms.Button{
						ID:                "button-reply",
						Class:             "max bold primary-container white-text thicc",
						Icon:              "reply",
						Text:              "Reply",
						OnClickActionName: m.OnClickReplyActionName,
						Disabled:          *m.ModalButtonsDisabled,
					},
				),
				app.Div().Class("space"),
			)
		}),
	)
}
