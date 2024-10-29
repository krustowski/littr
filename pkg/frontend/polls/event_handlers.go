package polls

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// onClickDelete()
func (c *Content) onClickDelete(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("delete", c.interactedPollKey)

	ctx.Dispatch(func(ctx app.Context) {
		c.deleteModalButtonsDisabled = true
	})
}

// onClickDeleteButton()
func (c *Content) onClickDeleteButton(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	ctx.Dispatch(func(ctx app.Context) {
		c.interactedPollKey = key
		c.deleteModalButtonsDisabled = false
		c.deletePollModalShow = true
	})
}

// onClickDismiss()
func (c *Content) onClickDismiss(ctx app.Context, e app.Event) {
	ctx.NewAction("dismiss")
}

// onClickLink()
func (c *Content) onClickLink(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()

	url := ctx.Page().URL()
	scheme := url.Scheme
	host := url.Host

	// write the link to browsers's clipboard
	navigator := app.Window().Get("navigator")
	if !navigator.IsNull() {
		clipboard := navigator.Get("clipboard")
		if !clipboard.IsNull() && !clipboard.IsUndefined() {
			clipboard.Call("writeText", scheme+"://"+host+"/polls/"+key)
		}
	}
	ctx.Navigate("/polls/" + key)
}

// onClickPollOption()
func (c *Content) onClickPollOption(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	option := ctx.JSSrc().Get("name").String()

	ctx.NewActionWithValue("vote", []string{key, option})

	ctx.Dispatch(func(ctx app.Context) {
		c.pollsButtonDisabled = true
	})
}

// onKeyDown()
func (c *Content) onKeyDown(ctx app.Context, e app.Event) {
	if e.Get("key").String() == "Escape" || e.Get("key").String() == "Esc" {
		ctx.NewAction("dismiss")
		return
	}
}

// onScroll()
func (c *Content) onScroll(ctx app.Context, e app.Event) {
	ctx.NewAction("scroll")
}
