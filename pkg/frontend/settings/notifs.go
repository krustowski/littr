package settings

import (
	"time"

	"go.vxn.dev/littr/pkg/frontend/common"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func (c *Content) checkPermission(ctx app.Context) bool {
	toast := common.Toast{AppContext: &ctx}

	// Notify user that their browser does not support notifications and therefore they cannot
	// use notifications service.
	if c.notificationPermission == app.NotificationNotSupported {
		toast.Text(common.ERR_NOTIF_UNSUPPORTED_BROWSER).Type(common.TTYPE_ERR).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			c.replyNotifOn = false
			c.subscribed = false
		})
		return false
	}

	// Request the permission on default when switch is toggled
	if (c.notificationPermission == app.NotificationDefault) ||
		(c.notificationPermission == app.NotificationDenied) {
		c.notificationPermission = ctx.Notifications().RequestPermission()
	}

	// fail on denied permission
	if c.notificationPermission != app.NotificationGranted {
		toast.Text(common.ERR_NOTIF_PERMISSION_DENIED).Type(common.TTYPE_ERR).Dispatch()

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

func (c *Content) createSubscription(ctx app.Context, tag string) {
	toast := common.Toast{AppContext: &ctx}

	var deviceSub models.Device

	// fetch the unique identificator for notifications and unsubscribe option
	uuid := ctx.DeviceID()

	vapidPubKey := c.VAPIDpublic
	if vapidPubKey == "" {
		// compiled into frontend/common/api.VapidPublicKey variable in wasm binary (see Dockerfile for more info)
		vapidPubKey = common.VapidPublicKey
	}

	// register the subscription
	sub, err := ctx.Notifications().Subscribe(vapidPubKey)
	if err != nil {
		toast.Text(common.ERR_SUBSCRIPTION_REQ_FAIL + err.Error()).Type(common.TTYPE_ERR).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false
			c.subscribed = false
		})
		return
	}

	// we need to convert type app.NotificationSubscription into webpush.Subscription
	webSub := models.Subscription{
		Endpoint: sub.Endpoint,
		Keys: models.Keys{
			Auth:   sub.Keys.Auth,
			P256dh: sub.Keys.P256dh,
		},
	}

	// compose the Device struct to be saved to the database
	deviceSub = models.Device{
		UUID:         uuid,
		TimeCreated:  time.Now(),
		Subscription: webSub,
		Tags: []string{
			tag,
		},
	}

	// send the registration to backend
	input := &common.CallInput{
		Method:      "POST",
		Url:         "/api/v1/users/" + c.user.Nickname + "/subscriptions",
		Data:        deviceSub,
		CallerID:    c.user.Nickname,
		PageNo:      0,
		HideReplies: false,
	}

	output := &common.Response{}

	if ok := common.FetchData(input, output); !ok {
		toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			c.subscribed = false
		})
		return
	}

	if output.Code != 201 && output.Code != 200 {
		toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
		return
	}

	devs := c.devices
	devs = append(devs, deviceSub)

	toast.Text(common.MSG_SUBSCRIPTION_REQ_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch()

	// Update the LocalStorage.
	common.SaveUser(&c.user, &ctx)

	// dispatch the good news to client
	ctx.Dispatch(func(ctx app.Context) {
		//c.user = user
		c.subscribed = true

		if tag == "mention" {
			c.subscription.Mentions = !c.subscription.Mentions
		} else if tag == "reply" {
			c.subscription.Replies = !c.subscription.Replies
		}

		c.thisDevice = deviceSub
		c.devices = devs
	})
}

func (c *Content) deleteSubscription(ctx app.Context) {
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
			Url:         "/api/v1/users/" + c.user.Nickname + "/subscriptions/" + ctx.DeviceID(),
			Data:        payload,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_CANNOT_REACH_BE).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				c.subscribed = true
				c.settingsButtonDisabled = false
			})
			return

		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		toast.Text(common.MSG_UNSUBSCRIBED_SUCCESS).Type(common.TTYPE_SUCCESS).Dispatch()

		ctx.Dispatch(func(ctx app.Context) {
			c.settingsButtonDisabled = false

			c.subscription.Mentions = false
			c.subscription.Replies = false

			c.subscribed = false
			c.thisDevice = models.Device{}
			c.devices = newDevs
		})
	})
}

func (c *Content) updateSubscriptionTag(ctx app.Context, tag string) {
	c.settingsButtonDisabled = true

	defer func() {
		c.settingsButtonDisabled = false
	}()

	var newDevs = make([]models.Device, 0)

	for _, dev := range c.devices {
		if dev.UUID == ctx.DeviceID() {
			if len(c.checkTags(dev.Tags, tag)) == 0 {
				continue
			}
			dev.Tags = c.checkTags(dev.Tags, tag)
		}
		newDevs = append(newDevs, dev)
	}

	payload := struct {
		Tags []string `json:"tags"`
	}{
		Tags: []string{tag},
	}

	toast := common.Toast{AppContext: &ctx}

	ctx.Async(func() {
		input := &common.CallInput{
			Method:      "PATCH",
			Url:         "/api/v1/users/" + c.user.Nickname + "/subscriptions/" + ctx.DeviceID(),
			Data:        payload.Tags,
			CallerID:    c.user.Nickname,
			PageNo:      0,
			HideReplies: false,
		}

		output := &common.Response{}

		if ok := common.FetchData(input, output); !ok {
			toast.Text(common.ERR_SUBSCRIPTION_UPDATE_FAIL).Type(common.TTYPE_ERR).Dispatch()

			ctx.Dispatch(func(ctx app.Context) {
				//c.subscribed = true
				c.settingsButtonDisabled = false
			})
			return
		}

		if output.Code != 200 {
			toast.Text(output.Message).Type(common.TTYPE_ERR).Dispatch()
			return
		}

		deviceSub := c.thisDevice
		deviceSub.Tags = c.checkTags(c.thisDevice.Tags, tag)

		toast.Text(common.MSG_SUBSCRIPTION_UPDATED).Type(common.TTYPE_SUCCESS).Dispatch()

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
	})
}
