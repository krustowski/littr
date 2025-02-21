package stats

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) onClickUserFlow(ctx app.Context, e app.Event) {
	key := ctx.JSSrc().Get("id").String()
	//c.buttonDisabled = true

	ctx.Navigate("/flow/users/" + key)
}

func (c *Content) onDismissToast(ctx app.Context, e app.Event) {
	c.toast.TText = ""
}

func (c *Content) onSearch(ctx app.Context, e app.Event) {
	val := ctx.JSSrc().Get("value").String()

	//if c.searchString == "" {
	//if val == "" {
	//	return
	//}

	if len(val) > 20 {
		return
	}

	ctx.NewActionWithValue("search", val)
}
