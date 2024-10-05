package post

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) Render() app.UI {
	toastColor := ""

	switch c.toast.TType {
	case "success":
		toastColor = "green10"
		break

	case "info":
		toastColor = "blue10"
		break

	default:
		toastColor = "red10"
	}

	return app.Main().Class("responsive").Body(
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("add flow post"),
				//app.P().Text("drop it, drop it"),
			),
		),

		// snackbar
		app.A().OnClick(c.onDismissToast).Body(
			app.If(c.toast.TText != "",
				app.Div().ID("snackbar").Class("snackbar white-text top active "+toastColor).Body(
					app.I().Text("error"),
					app.Span().Text(c.toast.TText),
				),
			),
		),

		// new post textarea
		app.Div().Class("field textarea label border extra deep-orange-text").Body(
			app.Textarea().Class("active").Name("newPost").OnChange(c.ValueTo(&c.newPost)).AutoFocus(true).ID("post-textarea").TabIndex(1),
			app.Label().Text("post content").Class("active deep-orange-text"),
		),
		/*app.Button().ID("post").Class("responsive deep-orange7 white-text bold").OnClick(c.onClick).Disabled(c.postButtonsDisabled).Body(
			app.If(c.postButtonsDisabled,
				app.Progress().Class("circle white-border small"),
			),
			app.Text("post text"),
		),*/

		// new fig input
		app.Div().Class("field border label extra deep-orange-text").Body(
			app.Input().ID("fig-upload").Class("active").Type("file").OnChange(c.ValueTo(&c.newFigLink)).OnInput(c.handleFigUpload).Accept("image/*").TabIndex(2),
			app.Input().Class("active").Type("text").Value(c.newFigFile).Disabled(true),
			app.Label().Text("image").Class("active deep-orange-text"),
			app.I().Text("image"),
		),
		app.Div().Class("row").Body(
			app.Button().ID("post").Class("max shrink center deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClick).Disabled(c.postButtonsDisabled).On("keydown", c.onKeyDown).TabIndex(3).Body(
				app.If(c.postButtonsDisabled,
					app.Progress().Class("circle white-border small"),
				),
				app.Text("send new post"),
			),
		),

		app.Div().Class("space"),

		// new poll header text
		app.Div().Class("row").Body(
			app.Div().Class("max padding").Body(
				app.H5().Text("add flow poll"),
				//app.P().Text("lmao gotem"),
			),
		),
		app.Div().Class("space"),

		// newx poll input area
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("poll-question").Type("text").OnChange(c.ValueTo(&c.pollQuestion)).Required(true).Class("active").MaxLength(50).TabIndex(4),
			app.Label().Text("question").Class("active deep-orange-text"),
		),
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("poll-option-i").Type("text").OnChange(c.ValueTo(&c.pollOptionI)).Required(true).Class("active").MaxLength(50).TabIndex(5),
			app.Label().Text("option one").Class("active deep-orange-text"),
		),
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("poll-option-ii").Type("text").OnChange(c.ValueTo(&c.pollOptionII)).Required(true).Class("active").MaxLength(50).TabIndex(6),
			app.Label().Text("option two").Class("active deep-orange-text"),
		),
		app.Div().Class("field border label deep-orange-text").Body(
			app.Input().ID("poll-option-iii").Type("text").OnChange(c.ValueTo(&c.pollOptionIII)).Required(false).Class("active").MaxLength(60).TabIndex(7),
			app.Label().Text("option three (optional)").Class("active deep-orange-text"),
		),
		app.Div().Class("row").Body(
			app.Button().ID("poll").Class("max shrink center deep-orange7 white-text bold").Style("border-radius", "8px").OnClick(c.onClick).Disabled(c.postButtonsDisabled).TabIndex(8).Body(
				app.If(c.postButtonsDisabled,
					app.Progress().Class("circle white-border small"),
				),
				app.Text("send new poll"),
			),
		),
		app.Div().Class("space"),
	)
}
