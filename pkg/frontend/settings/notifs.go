package settings

import (
	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func (c *Content) checkPermission(ctx app.Context, checked bool) bool {
	toast := common.Toast{AppContext: &ctx}

	// notify user that their browser does not support notifications and therefore they cannot
	// use notifications service
	if c.notificationPermission == app.NotificationNotSupported && checked {
		toast.Text("notifications are not supported in this browser").Type("error").Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			c.replyNotifOn = false
			c.subscribed = false
		})
		return false
	}

	// request the permission on default when switch is toggled
	if (c.notificationPermission == app.NotificationDefault && checked) ||
		(c.notificationPermission == app.NotificationDenied) {
		c.notificationPermission = ctx.Notifications().RequestPermission()
	}

	// fail on denied permission
	if c.notificationPermission != app.NotificationGranted {
		toast.Text("notification permission denied by user").Type("error").Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			c.replyNotifOn = false
			c.subscribed = false
		})
		return false
	}
	return true
}

func (c *Content) checkTags(tags []string, tag string) []string {
	// delete the tag if tags contain it
	if helpers.Contains(tags, tag) {
		newTags := []string{}
		for _, t := range tags {
			if t == tag {
				continue
			}
			newTags = append(newTags, t)
		}
		return newTags
	}

	// add the tag if missing
	return append(tags, tag)
}

func (c *Content) deleteSubscription(ctx app.Context, tag string) {
	toast := common.Toast{AppContext: &ctx}
	uuid := ctx.DeviceID()

	c.settingsButtonDisabled = true

	payload := struct {
		UUID string `json:"device_uuid"`
	}{
		UUID: uuid,
	}

	devs := c.devices
	newDevs := []models.Device{}
	for _, dev := range devs {
		if dev.UUID == ctx.DeviceID() {
			continue
		}
		newDevs = append(newDevs, dev)
	}

	ctx.Async(func() {
		input := &common.CallInput{
			Method:      "DELETE",
			Url:         "/api/v1/push/subscription/" + ctx.DeviceID(),
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text("cannot reach backend").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				c.subscribed = true
				c.settingsButtonDisabled = false
			})
			return

		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		toast.Text("successfully unsubscribed, notifications off").Type("success").Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false

			c.subscription.Mentions = false
			c.subscription.Replies = false

			c.subscribed = false
			c.thisDevice = models.Device{}
			c.devices = newDevs
		})
		return
	})
	return
}

func (c *Content) updateSubscriptionTag(ctx app.Context, tag string) {
	c.settingsButtonDisabled = true

	devs := c.devices
	newDevs := []models.Device{}
	for _, dev := range devs {
		if dev.UUID == ctx.DeviceID() {
			if len(c.checkTags(dev.Tags, tag)) == 0 {
				continue
			}
			dev.Tags = c.checkTags(dev.Tags, tag)
		}
		newDevs = append(newDevs, dev)
	}

	deviceSub := c.thisDevice

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		input := &common.CallInput{
			Method:      "PUT",
			Url:         "/api/v1/push/subscription/" + ctx.DeviceID() + "/" + tag,
			Data:        deviceSub,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text("failed to update the subscription, try again later").Type("error").Dispatch(c, dispatch)

			ctx.Dispatch(func(ctx app.Context) {
				//c.subscribed = true
				c.settingsButtonDisabled = false
			})
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type("error").Dispatch(c, dispatch)
			return
		}

		deviceSub.Tags = c.checkTags(c.thisDevice.Tags, tag)

		toast.Text("subscription updated").Type("success").Dispatch(c, dispatch)

		ctx.Dispatch(func(ctx app.Context) {
			if tag == "mention" {
				c.subscription.Mentions = !c.subscription.Mentions
			} else if tag == "reply" {
				c.subscription.Replies = !c.subscription.Replies
			}

			c.thisDevice = deviceSub
			c.devices = newDevs
			c.settingsButtonDisabled = false
		})
		return
	})
	return
}
