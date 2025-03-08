package settings

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/frontend/common"
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

func (c *Content) handleNotificationSwitchChange(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	type notifSubscription struct {
		Reply   bool
		Mention bool
	}

	var tag string

	subStates := notifSubscription{
		Reply:   c.subscription.Replies,
		Mention: c.subscription.Mentions,
	}

	subscribedCurrent := func() bool {
		if subStates.Reply || subStates.Mention {
			return true
		}

		return false
	}()

	switch key {
	case "reply-notif-switch":
		subStates.Reply = !subStates.Reply
		tag = "reply"
	case "mention-notif-switch":
		subStates.Mention = !subStates.Mention
		tag = "mention"
	}

	subscribedNew := func() bool {
		if subStates.Reply || subStates.Mention {
			return true
		}

		return false
	}()

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		// Unsubscribing.
		if subscribedCurrent && !subscribedNew {
			//
			c.deleteSubscription(ctx)
			return
		}

		// Subscribing.
		if !subscribedCurrent && subscribedNew {
			if !c.checkPermission(ctx) {
				return
			}

			c.createSubscription(ctx, tag)
			return
		}

		c.updateSubscriptionTag(ctx, tag)
		return
	})
}

func (c *Content) handleOptionSwitchChange(ctx app.Context, a app.Action) {
	key, ok := a.Value.(string)
	if !ok {
		return
	}

	ctx.Dispatch(func(ctx app.Context) {
		c.settingsButtonDisabled = true
	})

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		defer ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
		})

		var message string

		// See options.go.
		payload := c.prefillPayload()

		switch key {
		case "ui-mode-switch":
			message = common.MSG_UI_MODE_TOGGLE
			payload.UIMode = !payload.UIMode

		case "local-time-mode-switch":
			message = common.MSG_LOCAL_TIME_TOGGLE
			payload.LocalTimeMode = !payload.LocalTimeMode

		case "live-mode-switch":
			message = common.MSG_LIVE_MODE_TOGGLE
			payload.LiveMode = !payload.LiveMode

		case "private-mode-switch":
			message = common.MSG_PRIVATE_MODE_TOGGLE
			payload.Private = !payload.Private
		}

		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/options",
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		// Dispatch the good news to client.
		ctx.Dispatch(func(ctx app.Context) {
			c.updateOptions(payload)

			// Update the LocalStorage.
			common.SaveUser(&c.user, &ctx)
		})

		toast.Text(message).Type(common.TTYPE_SUCCESS).Dispatch()
	})
}
