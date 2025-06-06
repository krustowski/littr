// The post (new flow post, or new poll) view and view-controllers logic package.
package post

import (
	//"fmt"
	"encoding/base64"

	"go.vxn.dev/littr/pkg/frontend/common"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Content struct {
	app.Compo

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
}

func (c *Content) OnMount(ctx app.Context) {
	if app.IsServer {
		return
	}

	// Always focus the post textarea when the post.Content component is mounted.
	app.Window().Get("document").Call("getElementById", "post-textarea").Call("focus")

	//c.keyDownEventListener = app.Window().AddEventListener("keydown", c.onKeyDown)
	ctx.Handle("dismiss", c.handleDismiss)
	ctx.Handle("send-poll", c.handlePostPoll)
	ctx.Handle("send-post", c.handlePostPoll)

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
	if err := ctx.LocalStorage().Get("newPostDraft", &c.newPost); err != nil {
		return
	}
	if err := ctx.LocalStorage().Get("newPostFigFile", &c.newFigFile); err != nil {
		return
	}

	var data string
	if err := ctx.LocalStorage().Get("newPostFigData", &data); err != nil {
		return
	}

	c.newFigData, _ = base64.StdEncoding.DecodeString(data)
}
