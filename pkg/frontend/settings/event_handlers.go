package settings

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) onAvatarChange(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue("avatar-change", e.Get("target").Get("files").Index(0))
}
