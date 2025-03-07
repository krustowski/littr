package settings

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

// handleDismiss()
func (c *Content) handleDismiss(ctx app.Context, a app.Action) {
	ctx.Dispatch(func(ctx app.Context) {
		c.toast.TText = ""
		c.toast.TType = ""

		c.settingsButtonDisabled = false
		c.deleteAccountModalShow = false
		c.deleteSubscriptionModalShow = false
	})
}

func (c *Content) handleModalShow(ctx app.Context, a app.Action) {
	id, ok := a.Value.(string)
	if !ok {
		return
	}

	switch a.Name {
	case "user-delete-modal-show":
		ctx.Dispatch(func(ctx app.Context) {
			c.deleteAccountModalShow = true
			c.settingsButtonDisabled = true
		})

	case "subscription-delete-modal-show":
		ctx.Dispatch(func(ctx app.Context) {
			c.deleteSubscriptionModalShow = true
			c.settingsButtonDisabled = true
			c.interactedUUID = id
		})
	}

}
