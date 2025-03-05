package post

import (
	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// onDismissToast is a callback function that call a new action to handle that request for you.
func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	// Cast a new valueless action.
	ctx.NewAction("dismiss")
}

// onKeyDown is a callback function that takes care of the keyboard key UI controlling.
func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	// IS the key an Escape/Esc? Then the dismiss action call.
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}

	// Fetch the post textarea.
	textarea := app.Window().GetElementByID("post-textarea")

	// If null, we null.
	if textarea.IsNull() {
		return
	}

	// Otherwise utilize the CTRL-Enter combination and send the post in.
	if e.Get("ctrlKey").Bool() && e.Get("key").String() == "Enter" && len(textarea.Get("value").String()) != 0 {
		app.Window().GetElementByID("post").Call("click")
	}
}

func (c *Content) onTextareaBlur(ctx app.Context, e app.Event) {
	// Save a new post draft, if the focus on textarea is lost.
	ctx.LocalStorage().Set("newPostDraft", ctx.JSSrc().Get("value").String())
}

// handleFixUpload is a callbackfunction that takes care of the figure/image uploading.
// https://github.com/maxence-charriere/go-app/issues/882
func (c *Content) handleFigUpload(ctx app.Context, e app.Event) {
	// Fetch the first file in the row (index 0).
	file := e.Get("target").Get("files").Index(0)

	//log.Println("name", file.Get("name").String())
	//log.Println("size", file.Get("size").Int())
	//log.Println("type", file.Get("type").String())

	// Nasty way on how to disable buttons.
	c.postButtonsDisabled = true

	// Instantiate the toast.
	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.postButtonsDisabled = false
		})

		/*ctx.Defer(func(ctx app.Context) {
			c.postButtonsDisabled = false
		})*/

		var (
			data         []byte
			err          error
			processedImg *[]byte
		)

		// Read the figure/image data.
		data, err = common.ReadFile(file)
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		processedImg, err = common.ProcessImage(&data)
		if err != nil {
			toast.Text(err.Error()).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Cast the image ready message.
		toast.Text(common.MSG_IMAGE_READY).Type(common.TTYPE_INFO).Dispatch()

		// Load the image data to the Content structure.
		ctx.Dispatch(func(ctx app.Context) {
			c.newFigFile = file.Get("name").String()
			c.newFigData = *processedImg

			// Save the figure data in LS as a backup.
			ctx.LocalStorage().Set("newPostFigFile", file.Get("name").String())
			ctx.LocalStorage().Set("newPostFigData", *processedImg)
		})
	})
}
