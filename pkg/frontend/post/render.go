package post

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/atomic/atoms"
	"go.vxn.dev/littr/pkg/frontend/atomic/molecules"
)

func (c *Content) Render() app.UI {
	return app.Main().Class("responsive").Body(

		//
		//  New post
		//

		&atoms.PageHeading{
			Title: "new post",
		},

		&atoms.Textarea{
			ID:               "post-textarea",
			Class:            "field textarea label border extra primary-text thicc",
			ContentPointer:   &c.newPost,
			Name:             "newPost",
			LabelText:        "Content",
			OnBlurActionName: "blur-post",
		},

		&molecules.ImageInput{
			ImageData:            &c.newFigData,
			ImageFile:            &c.newFigFile,
			ImageLink:            &c.newFigLink,
			ButtonsDisabled:      &c.postButtonsDisabled,
			LocalStorageFileName: "newPostImageFile",
			LocalStorageDataName: "newPostImageData",
		},

		// New post button.
		&atoms.Button{
			ID:                "button-new-post",
			Class:             "max responsive shrink center primary-container white-text bold thicc",
			Icon:              "send",
			Text:              "Send",
			Disabled:          c.postButtonsDisabled,
			ShowProgress:      c.postButtonsDisabled,
			OnClickActionName: "send-post",
		},

		app.Div().Class("space"),

		//
		//  New poll
		//

		&atoms.PageHeading{
			Title: "new poll",
		},
		app.Div().Class("space"),

		// newx poll input area
		app.Div().Class("field border label primary-text").Style("border-radius", "8px").Body(
			app.Input().ID("poll-question").Type("text").OnChange(c.ValueTo(&c.pollQuestion)).Required(true).Class("active").MaxLength(50).TabIndex(4),
			app.Label().Text("Question").Class("active primary-text"),
		),
		app.Div().Class("field border label primary-text").Style("border-radius", "8px").Body(
			app.Input().ID("poll-option-i").Type("text").OnChange(c.ValueTo(&c.pollOptionI)).Required(true).Class("active").MaxLength(50).TabIndex(5),
			app.Label().Text("Option one").Class("active primary-text"),
		),
		app.Div().Class("field border label primary-text").Style("border-radius", "8px").Body(
			app.Input().ID("poll-option-ii").Type("text").OnChange(c.ValueTo(&c.pollOptionII)).Required(true).Class("active").MaxLength(50).TabIndex(6),
			app.Label().Text("Option two").Class("active primary-text"),
		),
		app.Div().Class("field border label primary-text").Style("border-radius", "8px").Body(
			app.Input().ID("poll-option-iii").Type("text").OnChange(c.ValueTo(&c.pollOptionIII)).Required(false).Class("active").MaxLength(60).TabIndex(7),
			app.Label().Text("Option three (optional)").Class("active primary-text"),
		),

		&atoms.Button{
			ID:                "button-new-poll",
			Class:             "max responsive shrink center primary-container white-text bold thicc",
			Icon:              "send",
			Text:              "Send",
			Disabled:          c.postButtonsDisabled,
			ShowProgress:      c.postButtonsDisabled,
			OnClickActionName: "send-poll",
			TabIndex:          8,
		},

		app.Div().Class("space"),
	)
}
