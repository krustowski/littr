// The post (new flow post, or new poll) view and view-controllers logic package.
package post

import (
	//"fmt"
	"encoding/base64"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Content struct {
	app.Compo

	postType string

	newPost    string
	newFigLink string
	newFigFile string
	newFigData []byte

	pollQuestion  string
	pollOptionI   string
	pollOptionII  string
	pollOptionIII string

	toast common.Toast

	postButtonsDisabled bool

	keyDownEventListener func()
}

func (c *Content) OnMount(ctx app.Context) {
	// Always focus the post textarea when the post.Content component is mounted.
	app.Window().Get("document").Call("getElementById", "post-textarea").Call("focus")

	//c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
	ctx.Handle("dismiss", c.handleDismiss)

	/*app.Window().Call("addEventListener", "keydown", app.FuncOf(func(this app.Value, args []app.Value) any {
		key := args[0].Get("key")

		ctx.NewActionWithKey("keydown", key)
	}))*/

	/*app.Window().Call("addEventListener", "beforeunload", app.FuncOf(func(this app.Value, args []app.Value) any {
		// https://developer.mozilla.org/en-US/docs/Web/API/Window/beforeunload_event
		if !args[0].IsNull() {
			args[0].Call("preventDefault")
		}

		fmt.Println("beforeunload event")

		var textareaValue string

		textarea := app.Window().GetElementByID("post-textarea")
		if !textarea.IsNull() {
			textareaValue = textarea.Get("value").String()
		}

		// Save the post draft to localStorage.
		if c.newPost != "" || textareaValue != "" {
			ctx.LocalStorage().Set("draftNewPost", textareaValue)
		}

		return nil
	}))*/

	/*app.Window().Call("addEventListener", "visibilitychange", app.FuncOf(func(this app.Value, args []app.Value) any {
		//event := args[0]

		if app.Window().Get("document").Get("visibilityState").String() == "hidden" {
			var textareaValue string

			textarea := app.Window().GetElementByID("post-textarea")
			if !textarea.IsNull() {
				textareaValue = textarea.Get("value").String()
			}

			// Save the post draft to localStorage.
			if textareaValue != "" {
				ctx.LocalStorage().Set("draftNewPost", textareaValue)
			}
		}
		return nil
	}))*/

	// Load the saved draft from localStorage.
	ctx.LocalStorage().Get("newPostDraft", &c.newPost)
	ctx.LocalStorage().Get("newPostFigFile", &c.newFigFile)

	var data string
	ctx.LocalStorage().Get("newPostFigData", &data)

	c.newFigData, _ = base64.StdEncoding.DecodeString(data)
}
